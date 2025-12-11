package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"spawner/api"
	"spawner/internal/config"
	"spawner/internal/game"
	"spawner/internal/updater"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	// 1. Setup Logging (File based for production)
	logFile, err := os.OpenFile("spawner.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %v", err))
	}
	defer logFile.Close()
	
	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

	// 2. Load Config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("Starting Spawner", "region", cfg.Region, "port", cfg.Port, "master_url", cfg.MasterURL)

	// 2.5 Ensure Game Server Files are Installed
	if err := updater.EnsureInstalled(cfg, logger); err != nil {
		logger.Error("Failed to ensure game server files", "error", err)
		os.Exit(1)
	}

	// 3. Initialize Game Manager
	manager := game.NewManager(cfg, logger)

	// 3.5 Restore previous instances
	if err := manager.RestoreInstances(); err != nil {
		logger.Error("Failed to restore instances", "error", err)
		// We continue, assuming the state was empty or corrupt, starting fresh.
	}

	// 4. Register with Master Server
	spawnerID := registerLoop(cfg, manager, logger)

	// 5. Start Heartbeat Loop
	go heartbeatLoop(spawnerID, cfg, manager, logger)

	// 6. Initialize Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	handler := api.NewHandler(manager, cfg, logger)
	handler.RegisterRoutes(router)

	// 7. Run Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Remove from Master Server on shutdown (optional, but good practice)
	// We can try to delete it, but with short context it might fail.
	// The heartbeat cleanup will handle it anyway.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	manager.Shutdown()

	logger.Info("Server exiting")
}

// registerLoop attempts to register this Spawner with the Master Server.
// It blocks until registration is successful, retrying every 5 seconds.
func registerLoop(cfg *config.Config, manager *game.Manager, logger *slog.Logger) int {
	port, _ := strconv.Atoi(cfg.Port)
	maxInstances := cfg.MaxGamePort - cfg.MinGamePort // Approximate capacity
	if maxInstances < 1 {
		maxInstances = 1
	}

	payload := map[string]interface{}{
		"region":            cfg.Region,
		"host":              cfg.Host,
		"port":              port,
		"max_instances":     maxInstances,
		"current_instances": len(manager.ListInstances()),
		"status":            "active",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/spawners", cfg.MasterURL)

	for {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			logger.Error("Failed to create request", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if cfg.MasterAPIKey != "" {
			req.Header.Set("X-API-Key", cfg.MasterAPIKey)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				var res map[string]int
				if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
					logger.Info("Registered with Master Server", "id", res["id"])
					return res["id"]
				}
			}
			logger.Warn("Failed to register", "status", resp.Status)
		} else {
			logger.Warn("Failed to connect to Master Server", "error", err)
		}

		time.Sleep(5 * time.Second)
	}
}

// heartbeatLoop sends periodic status updates to the Master Server.
func heartbeatLoop(id int, cfg *config.Config, manager *game.Manager, logger *slog.Logger) {
	url := fmt.Sprintf("%s/api/spawners/%d/heartbeat", cfg.MasterURL, id)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 5 * time.Second}

	for range ticker.C {
		maxInstances := cfg.MaxGamePort - cfg.MinGamePort
		
		// Collect System Metrics
		var cpuUsage float64
		if percentages, err := cpu.Percent(0, false); err == nil && len(percentages) > 0 {
			cpuUsage = percentages[0]
		}
		
		var memUsed, memTotal uint64
		if v, err := mem.VirtualMemory(); err == nil {
			memUsed = v.Used
			memTotal = v.Total
		}

		var diskUsed, diskTotal uint64
		// Use current directory for disk usage
		if d, err := disk.Usage("."); err == nil {
			diskUsed = d.Used
			diskTotal = d.Total
		}

		// Read dynamic game version
		currentGameVersion := ""
		versionFile := filepath.Join(cfg.GameInstallDir, "version.txt")
		if content, err := os.ReadFile(versionFile); err == nil {
			currentGameVersion = string(bytes.TrimSpace(content))
		}

		status := "active"
		if manager.IsBusy() {
			status = "Updating"
		}

		payload := map[string]interface{}{
			"current_instances": len(manager.ListInstances()),
			"max_instances":     maxInstances,
			"status":            status,
			"cpu_usage":         cpuUsage,
			"mem_used":          memUsed,
			"mem_total":         memTotal,
			"disk_used":         diskUsed,
			"disk_total":        diskTotal,
			"game_version":      currentGameVersion,
		}
		
		body, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			logger.Error("Failed to create heartbeat request", "error", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if cfg.MasterAPIKey != "" {
			req.Header.Set("X-API-Key", cfg.MasterAPIKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Heartbeat failed", "error", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Warn("Heartbeat returned non-200", "status", resp.Status)
			if resp.StatusCode == http.StatusNotFound {
				logger.Warn("Spawner not found on master, exiting to restart/re-register")
				os.Exit(1)
			}
		}
	}
}
