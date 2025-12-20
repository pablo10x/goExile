package main

import (
	"context"
	"errors"
	"flag" // Import the flag package
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"spawner/api"
	"spawner/internal/config"
	"spawner/internal/enrollment"
	"spawner/internal/game"
	"spawner/internal/updater"
	"spawner/internal/ws"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Define the application version
const appVersion = "1.0.0"

func main() {
	// 0. Handle command-line flags immediately
	versionFlag := flag.Bool("v", false, "Print version information and exit")
	flag.Parse() // Parse flags here so -v can be caught early

	if *versionFlag {
		fmt.Printf("Spawner Version: %s\n", appVersion)
		fmt.Println("Available commands:")
		flag.PrintDefaults() // Print default usage for all flags
		os.Exit(0)
	}

	// Customize usage output
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  The spawner connects to a Master Server to manage game instances.\n")
		fmt.Fprintf(os.Stderr, "  Configuration can be set via environment variables or command-line flags.\n")
		fmt.Fprintf(os.Stderr, "  Command-line flags take precedence over environment variables.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -key <enrollment_key>                   (Initial enrollment)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url http://localhost:8081 -sp 8000     (Run with specific Master and starting port)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -region us-east -max 10                 (Set region and max instances)\n", os.Args[0])
	}
	// Note: flag.Parse() is already called above, so explicit usage printing will be done after -v is handled.

	// 1. Setup Logging (File based for production)
	logFile, err := os.OpenFile("spawner.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %v", err))
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

	// 2. Load Config
	cfg, err := config.Load() // config.Load now calls flag.Parse() internally as well for non-version flags
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("Starting Spawner", "region", cfg.Region, "port", cfg.Port, "master_url", cfg.MasterURL)

	// 2.5 Handle Enrollment if enrollment key is provided
	if cfg.EnrollmentKey != "" {
		logger.Info("Enrollment key provided, initiating enrollment process...")
		result, err := enrollment.Enroll(cfg, logger)
		if err != nil {
			logger.Error("Enrollment failed", "error", err)
			os.Exit(1)
		}

		// Update config with the received API key
		cfg.MasterAPIKey = result.APIKey
		logger.Info("Enrollment complete, API key received", "spawner_id", result.ID)

		// Save the API key to .env for future runs
		if err := saveAPIKeyToEnv(result.APIKey); err != nil {
			logger.Warn("Failed to save API key to .env file", "error", err)
			logger.Info("Please manually set MASTER_API_KEY=" + result.APIKey + " in your .env file")
		} else {
			logger.Info("API key saved to .env file for future runs")
		}
	}

	// 2.6 Ensure Game Server Files are Installed
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

// saveAPIKeyToEnv saves the API key to the .env file
func saveAPIKeyToEnv(apiKey string) error {
	envFile := ".env"

	// Read existing content
	content, err := os.ReadFile(envFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if MASTER_API_KEY already exists
	lines := []string{}
	found := false
	if len(content) > 0 {
		for _, line := range splitLines(string(content)) {
			if len(line) >= 14 && line[:14] == "MASTER_API_KEY" {
				lines = append(lines, "MASTER_API_KEY="+apiKey)
				found = true
			} else {
				lines = append(lines, line)
			}
		}
	}

	if !found {
		lines = append(lines, "MASTER_API_KEY="+apiKey)
	}

	// Write back
	newContent := ""
	for i, line := range lines {
		newContent += line
		if i < len(lines)-1 {
			newContent += "\n"
		}
	}

	return os.WriteFile(envFile, []byte(newContent), 0644)
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
