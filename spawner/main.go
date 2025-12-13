package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"spawner/api"
	"spawner/internal/config"
	"spawner/internal/game"
	"spawner/internal/updater"
	"spawner/internal/ws"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	// 4. Start WebSocket Client (Handles registration and heartbeat)
	wsClient := ws.NewClient(cfg, manager, logger)
	go wsClient.Start()

	// 6. Initialize Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Security: API Key Middleware
	router.Use(func(c *gin.Context) {
		// Public endpoints
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		if cfg.MasterAPIKey != "" {
			if c.GetHeader("X-API-Key") != cfg.MasterAPIKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		} else {
			// In Production, this should be fatal, but for now we warn
			// logger.Warn("Running without API Key auth protection!")
		}
		c.Next()
	})

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
