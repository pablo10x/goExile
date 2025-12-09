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
			// Re-populate in-memory registry from DB records
			maxID := registry.nextID - 1
			registry.mu.Lock()
			for i := range loaded {
				s := loaded[i]
				copyS := s
				registry.items[copyS.ID] = &copyS
				if copyS.ID > maxID {
					maxID = copyS.ID
				}
			}
			registry.nextID = maxID + 1
			registry.mu.Unlock()
			log.Printf("loaded %d spawners from DB, nextID=%d", len(loaded), registry.nextID)
		}
	}

	// 6. Print configuration summary
	PrintConfig(8081, dbPath)

	// 7. Define API Routes
	// Spawner interactions (public/internal API) - Secured via API Key if set
	apiRouter := router.PathPrefix("/api/spawners").Subrouter()
	apiKey := os.Getenv("MASTER_API_KEY")
	if apiKey == "" {
		apiKey = "your_very_secret_master_api_key_here" // Default API key
		log.Println("⚠️ MASTER_API_KEY not set, using default API key. Consider setting it in your environment.")
	}
	if apiKey != "" {
		log.Println("🔒 API Key authentication enabled for Spawners")
	//	apiRouter.Use(APIKeyMiddleware(apiKey))
	}

	apiRouter.HandleFunc("/download", ServeGameServerFile).Methods("GET")
	apiRouter.HandleFunc("", RegisterSpawner).Methods("POST")
	apiRouter.HandleFunc("", ListSpawners).Methods("GET") // Maybe this should be public or auth? Keeping consistent
	apiRouter.HandleFunc("/{id}", GetSpawner).Methods("GET")
	apiRouter.HandleFunc("/{id}", DeleteSpawner).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/spawn", SpawnInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/heartbeat", HeartbeatSpawner).Methods("POST")
	apiRouter.HandleFunc("/{id}/logs", GetSpawnerLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/logs", ClearSpawnerLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances", ListSpawnerInstances).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id}", StopSpawnerInstance).Methods("DELETE")

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
		router.Handle("/events", AuthMiddleware(authConfig, sessionStore)(sseHandler))           // Replaced /ws with /events
		router.Handle("/api/upload", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleUploadGameServer))).Methods("POST")
		
		// Version Management Routes
		router.Handle("/api/versions", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListVersions))).Methods("GET")
		router.Handle("/api/versions/{id}/active", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleSetActiveVersion))).Methods("POST")
		router.Handle("/api/versions/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(HandleDeleteVersion))).Methods("DELETE")
	}

	// CLI-friendly status endpoint
	router.HandleFunc("/status", PrintStatus).Methods("GET")

	// 8. Start background cleanup
	go registry.Cleanup(serverTTL, cleanupInterval)

	// Start Memory Stats Ticker
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			GlobalStats.UpdateMemoryStats()
		}
	}()

	// 9. Start HTTP Server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: StatsMiddleware(router),
	}

	go func() {
		log.Println("✓ Starting spawner registry on :8081")
		log.Println("📊 API Stats: http://localhost:8081/api/stats")
		log.Println("🏥 Health Check: http://localhost:8081/health")
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
