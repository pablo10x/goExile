// Package main provides a lightweight HTTP-based spawner registry.
//
// The registry acts as the central authority for tracking active Spawners.
// It exposes a REST API for:
// - Spawner registration and heartbeats
// - Dashboard monitoring (via WebSocket)
// - Spawning new game instances (proxying to Spawners)
package main

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Configuration constants used by the registry.
const (
	// maxBodySize is the maximum size accepted for request bodies.
	maxBodySize = 1 << 20 // 1MB

	// serverTTL defines how long a spawner is considered alive since its
	// last heartbeat. Spawners older than this are removed by cleanup.
	serverTTL = 60 * time.Second

	// cleanupInterval is how frequently the cleanup loop runs.
	cleanupInterval = 30 * time.Second

	// healthCheckInterval is how frequently the master server pings spawners for health checks.
	healthCheckInterval = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

// run initializes and starts the registry server.
// It handles database connection, route setup, background cleanup tasks,
// and graceful shutdown.
func run() error {
	// 1. Print startup banner
	PrintBanner()

	// 2. Initialize authentication & session management
	authConfig := GetAuthConfig()
	sessionStore := NewSessionStore()
	go sessionStore.CleanupExpiredSessions()

	// 3. Initialize SSE hub for real-time dashboard updates
	sseHub := NewSSEHub()
	go sseHub.Run()
	
	// 4. Initialize Router
	router := mux.NewRouter()
	
	// 5. Initialize Database (optional persistence)
	// If DB_PATH is not set, defaults to "database/registry.db".
	// Falls back to in-memory mode if DB connection fails.
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "database/registry.db"
	}
		var err error
	port := "8081"
	
	dbConn, err = InitDB(dbPath)
	if err != nil {
		log.Printf("warning: failed to init DB (%s): %v — continuing without persistence", dbPath, err)
		GlobalStats.SetDBConnected(false)
	} else {
		GlobalStats.SetDBConnected(true)
		// Load existing spawners from DB to restore state
		loaded, err := LoadSpawners(dbConn)
		if err != nil {
			log.Printf("warning: failed to load spawners from DB: %v", err)
		} else {
			// Re-populate in-memory registry from DB records, checking health
			maxID := registry.nextID - 1
			registry.mu.Lock()

			validCount := 0
			// client := &http.Client{Timeout: 2 * time.Second} // No longer needed for proactive checks

			for i := range loaded {
				s := loaded[i]

				// Health check removed as per user request (rely on heartbeats)

				copyS := s
				registry.items[copyS.ID] = &copyS
				if copyS.ID > maxID {
					maxID = copyS.ID
				}
				validCount++
			}
			registry.nextID = maxID + 1
			registry.mu.Unlock()
			log.Printf("Startup: loaded %d spawners from DB.", validCount)
		}
	}

	// 6. Print configuration summary
	PrintConfig(port, dbPath)

	// 7. Define API Routes
	// Spawner interactions (public/internal API) - Secured via API Key if set
	apiRouter := router.PathPrefix("/api/spawners").Subrouter()

	apiKey := os.Getenv("MASTER_API_KEY")
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"

    // Start WS Manager
    go GlobalWSManager.Run()

	if isProduction {
		if apiKey == "" {
			log.Fatal("FATAL: MASTER_API_KEY must be set in production mode")
		}
		log.Println("🔒 API Key authentication enabled for Spawners (Production)")
	} else {
		if apiKey == "" {
			apiKey = "dev_master_key" // Default for development
			log.Println("⚠️  MASTER_API_KEY not set, using default 'dev_master_key'")
		}
		log.Println("🔒 API Key authentication enabled for Spawners (Dev)")
	}

	// Always enforce Unified Auth (API Key OR Session)
	apiRouter.Use(UnifiedAuthMiddleware(apiKey, authConfig, sessionStore))

	apiRouter.HandleFunc("/ws", GlobalWSManager.HandleWS) // WebSocket Endpoint
	apiRouter.HandleFunc("/download", ServeGameServerFile).Methods("GET", "HEAD")
	apiRouter.HandleFunc("", RegisterSpawner).Methods("POST")
	apiRouter.HandleFunc("", ListSpawners).Methods("GET") // Maybe this should be public or auth? Keeping consistent
	apiRouter.HandleFunc("/{id}", GetSpawner).Methods("GET")
	apiRouter.HandleFunc("/{id}", DeleteSpawner).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/spawn", SpawnInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/heartbeat", HeartbeatSpawner).Methods("POST")
	apiRouter.HandleFunc("/{id}/logs", GetSpawnerLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/logs", ClearSpawnerLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances", ListSpawnerInstances).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", GetInstanceLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", ClearInstanceLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats", GetInstanceStats).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/start", StartSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stop", StopSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restart", RestartSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/update", UpdateSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/rename", RenameSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}", RemoveSpawnerInstance).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup", BackupSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restore", RestoreSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backups", ListSpawnerBackups).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup/delete", DeleteSpawnerBackup).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats/history", GetInstanceHistory).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/history", GetInstanceHistoryActions).Methods("GET")
	apiRouter.HandleFunc("/{id}/update-template", UpdateSpawnerTemplate).Methods("POST")

	// Liveness check
	router.HandleFunc("/health", Health).Methods("GET")

	// Authentication endpoints
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		HandleLogin(w, r, authConfig, sessionStore)
	}).Methods("GET", "POST")

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		HandleLogout(w, r, sessionStore)
	}).Methods("GET")

	// Dashboard & UI endpoints (Protected by AuthMiddleware in dev/prod)
	statsHandler := http.HandlerFunc(StatsAPI)
	errorsAPIHandler := http.HandlerFunc(ErrorsAPI) // New API handler
	sseHandler := http.HandlerFunc(sseHub.HandleSSE)

	if authConfig.Enabled {
		router.Handle("/api/stats", AuthMiddleware(authConfig, sessionStore)(statsHandler))
		router.Handle("/api/errors", AuthMiddleware(authConfig, sessionStore)(errorsAPIHandler)).Methods("GET")
		router.Handle("/api/errors", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ClearErrorsAPI))).Methods("DELETE")
		router.Handle("/events", AuthMiddleware(authConfig, sessionStore)(sseHandler)) // Replaced /ws with /events
		router.Handle("/api/upload", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleUploadGameServer))).Methods("POST")

		// Version Management Routes
		router.Handle("/api/versions", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListVersions))).Methods("GET")
		router.Handle("/api/versions/{id}/active", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleSetActiveVersion))).Methods("POST")
		router.Handle("/api/versions/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleDeleteVersion))).Methods("DELETE")

		// Configuration Management Routes
		router.Handle("/api/config", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetAllConfigHandler))).Methods("GET")
		router.Handle("/api/config", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateConfigHandler))).Methods("POST")
		router.Handle("/api/config/{category}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetConfigByCategoryHandler))).Methods("GET")
		router.Handle("/api/config/{key}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateConfigHandler))).Methods("PUT")
		router.Handle("/api/config/key/{key}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetConfigByKeyHandler))).Methods("GET")
	}

	// CLI-friendly status endpoint
	router.HandleFunc("/status", PrintStatus).Methods("GET")

	// 8. Start background cleanup
	go registry.Cleanup(serverTTL, cleanupInterval)

	// 9. Start proactive health checks
	// go ProactiveHealthCheck(healthCheckInterval)

	// Start Memory Stats Ticker
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			GlobalStats.UpdateMemoryStats()
		}
	}()

	// 10. Start HTTP Server
srv := &http.Server{
    Addr:    fmt.Sprintf(":%s", port),
    Handler: StatsMiddleware(router),
}

	go func() {
    log.Printf("✓ Starting spawner registry on :%s", port)

    apiURL := fmt.Sprintf("📊 API Stats: http://localhost:%s/api/stats", port)
    log.Println(apiURL)

    healthURL := fmt.Sprintf("🏥 Health Check: http://localhost:%s/health", port)
    log.Println(healthURL)

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("listen: %v", err)
    }
}()



	// 10. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received, shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
	if dbConn != nil {
		log.Println("closing database connection")
		dbConn.Close()
	}
	log.Println("server stopped")
	return nil
}
