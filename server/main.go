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
	"github.com/joho/godotenv"
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
	// Load .env file silently
	_ = godotenv.Load()

	// Print startup banner
	PrintBanner()

	fmt.Println()
	fmt.Printf("  %s%s▸ Initializing Services%s\n", "\033[1m", "\033[36m", "\033[0m")
	fmt.Printf("  %s────────────────────────────────────────%s\n", "\033[2m", "\033[0m")

	// Initialize authentication & session management
	authConfig := GetAuthConfig()
	sessionStore := NewSessionStore(authConfig.IsProduction)
	go sessionStore.CleanupExpiredSessions()
	PrintSection("Authentication", "ready", true)

	// Initialize Enrollment Manager for spawner enrollment
	InitializeEnrollmentManager()
	PrintSection("Enrollment Manager", "ready", true)

	// Initialize SSE hub for real-time dashboard updates
	sseHub := NewSSEHub()
	go sseHub.Run()
	PrintSection("SSE Hub", "ready", true)

	// Initialize Router
	router := mux.NewRouter()

	// Initialize Database (PostgreSQL)
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		PrintSection("Database", "disabled (no DB_DSN)", false)
	} else {
		var err error
		dbConn, err = InitDB(dbDSN)
		if err != nil {
			PrintSection("Database", "failed", false)
			PrintSubItem(err.Error())
			GlobalStats.SetDBConnected(false)
		} else {
			GlobalStats.SetDBConnected(true)
			// Load existing spawners from DB to restore state
			loaded, _ := LoadSpawners(dbConn)
			if len(loaded) > 0 {
				maxID := registry.nextID - 1
				registry.mu.Lock()
				for i := range loaded {
					s := loaded[i]
					copyS := s
					// Reconcile status: if marked online in DB but not actually connected via WS, set to offline
					if copyS.Status == "Online" && !GlobalWSManager.IsClientConnected(copyS.ID) {
						log.Printf("INFO: Spawner ID %d was marked 'Online' in DB but not connected to WebSocket. Marking 'Offline'.", copyS.ID)
						copyS.Status = "Offline"
						// Optionally, persist this change to DB immediately, but WSManager will handle it on reconnect/disconnect
						// For now, only update in-memory to reflect reality for the Dashboard
					}
					registry.items[copyS.ID] = &copyS
					if copyS.ID > maxID {
						maxID = copyS.ID
					}
				}
				registry.nextID = maxID + 1
				registry.mu.Unlock()
				PrintSection("Database", "connected", true)
				PrintSubItem(fmt.Sprintf("Loaded %d spawners", len(loaded)))
			} else {
				PrintSection("Database", "connected", true)
			}
		}

		// Initialize Read-Only Database (Optional but recommended)
		roDSN := os.Getenv("READONLY_DB_DSN")
		if roDSN != "" {
			readOnlyConn, err = InitReadOnlyDB(roDSN)
			if err != nil {
				PrintSection("Read-Only DB", "failed", false)
				PrintSubItem(err.Error())
			} else {
				PrintSection("Read-Only DB", "connected", true)
			}
		} else {
			// If not set, use main connection but fallback to regex validation
			readOnlyConn = dbConn
			if dbConn != nil {
				PrintSection("Read-Only DB", "using primary (fallback)", true)
			}
		}
	}

	// Initialize Firebase Remote Config
	_ = InitFirebase()
	if firebaseManager != nil && firebaseManager.connected {
		PrintSection("Firebase", "connected", true)
		PrintSubItem(fmt.Sprintf("Project: %s", firebaseManager.projectID))
	} else {
		PrintSection("Firebase", "disabled", false)
	}

	port := "8081"

	// Define API Routes

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
	} else {
		if apiKey == "" {
			apiKey = "dev_master_key" // Default for development
		}
	}

	// Always enforce Unified Auth (API Key OR Session)
	apiRouter.Use(UnifiedAuthMiddleware(apiKey, authConfig, sessionStore))

	apiRouter.HandleFunc("/ws", GlobalWSManager.HandleWS) // WebSocket Endpoint
	apiRouter.HandleFunc("/download", ServeGameServerFile).Methods("GET", "HEAD")
	// apiRouter.HandleFunc("", RegisterSpawner).Methods("POST") // Disabled to enforce enrollment flow
	apiRouter.HandleFunc("", ListSpawners).Methods("GET")     // Maybe this should be public or auth? Keeping consistent
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
	}).Methods("POST")

	router.HandleFunc("/login/2fa", func(w http.ResponseWriter, r *http.Request) {
		Handle2FAVerify(w, r, authConfig, sessionStore)
	}).Methods("POST")

	router.HandleFunc("/login/email", func(w http.ResponseWriter, r *http.Request) {
		HandleEmailVerify(w, r, authConfig, sessionStore)
	}).Methods("POST")

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

		// Database Management Routes
		router.Handle("/api/database/tables", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListTablesHandler))).Methods("GET")
		router.Handle("/api/database/counts", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetTableCountsHandler))).Methods("GET")
		router.Handle("/api/database/backup", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(BackupDatabaseHandler))).Methods("POST")

		router.Handle("/api/database/overview", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetDatabaseOverviewHandler))).Methods("GET")
		router.Handle("/api/database/table/{table}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetTableDataHandler))).Methods("GET")
		router.Handle("/api/database/table/{table}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(InsertTableRowHandler))).Methods("POST")
		router.Handle("/api/database/table/{table}/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateTableRowHandler))).Methods("PUT")
		router.Handle("/api/database/table/{table}/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteTableRowHandler))).Methods("DELETE")
		router.Handle("/api/database/config", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetPostgresConfigHandler))).Methods("GET")
		router.Handle("/api/database/config", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdatePostgresConfigHandler))).Methods("PUT")
		router.Handle("/api/database/config/restart", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(RestartPostgresHandler))).Methods("POST")

		// Introspection & SQL
		router.Handle("/api/database/schemas", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListSchemasHandler))).Methods("GET")
		router.Handle("/api/database/schemas", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateSchemaHandler))).Methods("POST")
		router.Handle("/api/database/schemas/{name}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteSchemaHandler))).Methods("DELETE")
		router.Handle("/api/database/all-tables", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetAllTablesHandler))).Methods("GET")
		router.Handle("/api/database/tables", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListTablesBySchemaHandler))).Methods("GET")
		router.Handle("/api/database/tables/create", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateTableHandler))).Methods("POST")
		router.Handle("/api/database/tables/{table}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DropTableHandler))).Methods("DELETE")
		router.Handle("/api/database/tables/{table}/alter", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(AlterTableHandler))).Methods("POST")
		router.Handle("/api/database/columns", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListColumnsHandler))).Methods("GET")
		router.Handle("/api/database/sql", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ExecuteSQLHandler))).Methods("POST")
		router.Handle("/api/database/debug/tables", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DebugListAllTablesHandler))).Methods("GET")

		// Roles
		router.Handle("/api/database/roles", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListRolesHandler))).Methods("GET")
		router.Handle("/api/database/roles", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateRoleHandler))).Methods("POST")
		router.Handle("/api/database/roles/{name}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteRoleHandler))).Methods("DELETE")

		// Functions
		router.Handle("/api/database/functions", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListFunctionsHandler))).Methods("GET")
		router.Handle("/api/database/functions/details", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetFunctionHandler))).Methods("GET")
		router.Handle("/api/database/functions", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateFunctionHandler))).Methods("POST")
		router.Handle("/api/database/functions", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateFunctionHandler))).Methods("PUT")
		router.Handle("/api/database/functions/delete", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteFunctionHandler))).Methods("POST")
		router.Handle("/api/database/functions/execute", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ExecuteFunctionHandler))).Methods("POST")

		router.Handle("/api/database/backups", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateInternalBackupHandler))).Methods("POST")
		router.Handle("/api/database/backups", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListInternalBackupsHandler))).Methods("GET")
		router.Handle("/api/database/backups/{filename}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DownloadInternalBackupHandler))).Methods("GET")
		router.Handle("/api/database/backups/{filename}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteInternalBackupHandler))).Methods("DELETE")
		router.Handle("/api/database/restore", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(RestoreInternalBackupHandler))).Methods("POST")

		// System Management
		router.Handle("/api/restart", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(RestartServerHandler))).Methods("POST")

		// Performance Metrics API
		router.Handle("/api/metrics", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetAllMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/master", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetMasterMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/spawners", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetSpawnerMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/database", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetDatabaseMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/network", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetNetworkMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/gc", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ForceGCHandler))).Methods("POST")
		router.Handle("/api/metrics/memory/free", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(FreeMemoryHandler))).Methods("POST")

		// Enrollment Key Management (Dashboard authenticated endpoints)
		router.Handle("/api/enrollment/generate", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GenerateEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/keys", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListEnrollmentKeysHandler))).Methods("GET")
		router.Handle("/api/enrollment/revoke", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(RevokeEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/status", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetEnrollmentKeyStatusHandler))).Methods("POST")

		// Firebase Remote Config Routes
		router.Handle("/api/config/firebase/status", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetFirebaseStatusHandler))).Methods("GET")
		router.Handle("/api/config/firebase/configs", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetFirebaseConfigsHandler))).Methods("GET")
		router.Handle("/api/config/firebase/sync", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(SyncFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateFirebaseConfigHandler))).Methods("PUT")
		router.Handle("/api/config/firebase/parameter", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteFirebaseConfigHandler))).Methods("DELETE")
		router.Handle("/api/config/firebase/publish", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(PublishFirebaseConfigHandler))).Methods("POST")

		// Notes & Todos API
		router.Handle("/api/notes", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListNotesHandler))).Methods("GET")
		router.Handle("/api/notes", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateNoteHandler))).Methods("POST")
		router.Handle("/api/notes/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateNoteHandler))).Methods("PUT")
		router.Handle("/api/notes/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteNoteHandler))).Methods("DELETE")

		router.Handle("/api/todos", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListTodosHandler))).Methods("GET")
		router.Handle("/api/todos", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateTodoHandler))).Methods("POST")
		router.Handle("/api/todos/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateTodoHandler))).Methods("PUT")
		router.Handle("/api/todos/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteTodoHandler))).Methods("DELETE")

		// Logging
		router.Handle("/api/logs", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListSystemLogsHandler))).Methods("GET")
		router.Handle("/api/logs/counts", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(GetSystemLogCountsHandler))).Methods("GET")
		router.Handle("/api/logs", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ClearSystemLogsHandler))).Methods("DELETE")
		router.Handle("/api/logs/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteSystemLogHandler))).Methods("DELETE")

		// RedEye Security System
		router.Handle("/api/redeye/rules", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListRedEyeRulesHandler))).Methods("GET")
		router.Handle("/api/redeye/rules", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(CreateRedEyeRuleHandler))).Methods("POST")
		router.Handle("/api/redeye/rules/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(UpdateRedEyeRuleHandler))).Methods("PUT")
		router.Handle("/api/redeye/rules/{id}", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(DeleteRedEyeRuleHandler))).Methods("DELETE")
		router.Handle("/api/redeye/logs", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListRedEyeLogsHandler))).Methods("GET")
		router.Handle("/api/redeye/logs", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ClearRedEyeLogsHandler))).Methods("DELETE")
		
		// RedEye Anti-Cheat
		router.Handle("/api/redeye/anticheat/report", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ReportAnticheatEventHandler))).Methods("POST")
		router.Handle("/api/redeye/anticheat/events", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(ListAnticheatEventsHandler))).Methods("GET")

		// AI Bot API
		router.Handle("/api/ai/chat", AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(AIChatHandler))).Methods("POST")
	}

	// Enrollment endpoints (public - enrollment key IS the auth)
	router.HandleFunc("/api/enrollment/register", RegisterSpawnerWithKeyHandler).Methods("POST")
	router.HandleFunc("/api/enrollment/validate", ValidateEnrollmentKeyHandler).Methods("POST")

	// CLI-friendly status endpoint
	router.HandleFunc("/status", PrintStatus).Methods("GET")

	// Start proactive health checks
	// go ProactiveHealthCheck(healthCheckInterval)

	// Start Stats Ticker (Memory & DB)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			GlobalStats.UpdateMemoryStats()
			if dbConn != nil {
				GlobalStats.UpdateDBStats(dbConn.Stats())
				// Advanced Stats
				advStats, err := GetAdvancedDBStats(dbConn)
				if err == nil {
					GlobalStats.UpdateAdvancedDBStats(advStats)
				} else {
					log.Printf("Failed to get advanced DB stats: %v", err)
				}
			}
		}
	}()

	// 10. Start HTTP Server
	// If SERVER_HOST env is set, use it (e.g. "0.0.0.0" for Docker), otherwise default to "127.0.0.1"
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "127.0.0.1"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", serverHost, port),
		Handler: SecurityHeadersMiddleware(RedEyeMiddleware(StatsMiddleware(router))),
	}

	go func() {
		PrintStartupComplete(port)
		// Warn if not binding to localhost
		if serverHost != "127.0.0.1" && serverHost != "localhost" {
			fmt.Printf("  %s%s⚠️  WARNING: Server is listening on %s (Potentially public)%s\n", "\033[1m", "\033[33m", serverHost, "\033[0m")
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 10. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Printf("\n  %s●%s Shutting down gracefully...%s\n", "\033[33m", "\033[0m", "\033[0m")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
	if dbConn != nil {
		dbConn.Close()
	}
	log.Println("server stopped")
	return nil
}
