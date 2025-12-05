// Package main provides a lightweight HTTP-based game server registry.
//
// The registry exposes a small REST API that allows game server instances
// (for example, Unity servers) to register themselves, send periodic
// heartbeats to indicate liveness, and be discovered by players. The
// implementation uses an in-memory registry with an optional SQLite
// persistence layer. A background cleanup goroutine removes servers
// that have not heartbeated within the configured TTL.
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

// Configuration constants used by the registry. These constants keep the
// runtime behaviour simple and predictable for small deployments.
const (
	// maxBodySize is the maximum size accepted for request bodies. This
	// protects the server from large or inefficient payloads.
	maxBodySize = 1 << 20 // 1MB

	// maxIDValue provides a sanity limit for parsed ID values.
	maxIDValue = 1000000

	// serverTTL defines how long a server is considered alive since its
	// last heartbeat. Servers older than this are removed by cleanup.
	serverTTL = 60 * time.Second

	// cleanupInterval is how frequently the cleanup loop runs.
	cleanupInterval = 30 * time.Second
)

// (ErrorResponse moved to models.go)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

// run creates the router, wires handlers, optionally initializes persistence,
// and starts the HTTP server. The function blocks until the process receives
// an interrupt signal, at which point it performs a graceful shutdown.
func run() error {
	// Print banner
	PrintBanner()

	router := mux.NewRouter()

	// Initialize DB (optional). Path can be configured with DB_PATH env var.
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "registry.db"
	}
	var err error
	dbConn, err = InitDB(dbPath)
	if err != nil {
		log.Printf("warning: failed to init DB (%s): %v — continuing without persistence", dbPath, err)
		GlobalStats.SetDBConnected(false)
	} else {
		GlobalStats.SetDBConnected(true)
		// Load existing servers from DB into in-memory registry
		loaded, err := LoadServers(dbConn)
		if err != nil {
			log.Printf("warning: failed to load servers from DB: %v", err)
		} else {
			// populate registry and ensure nextID is greater than any existing id
			maxID := registry.nextID - 1
			registry.mu.Lock()
			for i := range loaded {
				s := loaded[i]
				// copy into map
				copyS := s
				registry.items[copyS.ID] = &copyS
				if copyS.ID > maxID {
					maxID = copyS.ID
				}
			}
			registry.nextID = maxID + 1
			registry.mu.Unlock()
			log.Printf("loaded %d servers from DB, nextID=%d", len(loaded), registry.nextID)
		}
	}

	// Print configuration
	PrintConfig(8081, dbPath)

	// Server registry routes
	router.HandleFunc("/api/servers", RegisterServer).Methods("POST")
	router.HandleFunc("/api/servers", ListServers).Methods("GET")
	router.HandleFunc("/api/servers/{id}", GetServer).Methods("GET")
	router.HandleFunc("/api/servers/{id}", DeleteServer).Methods("DELETE")
	router.HandleFunc("/api/servers/{id}/heartbeat", HeartbeatServer).Methods("POST")

	// Health check
	router.HandleFunc("/health", Health).Methods("GET")

	// Dashboard endpoints
	router.HandleFunc("/", DashboardPage).Methods("GET")
	router.HandleFunc("/dashboard", DashboardPage).Methods("GET")
	router.HandleFunc("/api/stats", StatsAPI).Methods("GET")

	// Dashboard status endpoint (text format)
	router.HandleFunc("/status", PrintStatus).Methods("GET")

	// Start cleanup goroutine
	go registry.Cleanup(serverTTL, cleanupInterval)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	// Run server in background so we can handle graceful shutdown and close DB
	go func() {
		log.Println("✓ Starting game server registry on :8081")
		log.Println("🌐 Web Dashboard: http://localhost:8081")
		log.Println("📊 API Stats: http://localhost:8081/api/stats")
		log.Println("🏥 Health Check: http://localhost:8081/health")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Wait for termination signal
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
