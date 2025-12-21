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

	"exile/server/auth"
	"exile/server/config"
	"exile/server/database"
	"exile/server/enrollment"
	"exile/server/handlers"
	"exile/server/middleware"
	"exile/server/redeye"
	"exile/server/registry"
	"exile/server/sse"
	"exile/server/utils"
	"exile/server/ws"
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
	utils.PrintBanner()

	fmt.Println()
	fmt.Printf("  %s%s▸ Initializing Services%s\n", "\033[1m", "\033[36m", "\033[0m")
	fmt.Printf("  %s────────────────────────────────────────%s\n", "\033[2m", "\033[0m")

	// Initialize authentication & session management
	authConfig := auth.GetAuthConfig()
	sessionStore := auth.NewSessionStore(authConfig.IsProduction)
	go sessionStore.CleanupExpiredSessions()
	utils.PrintSection("Authentication", "ready", true)

	// Initialize Enrollment Manager for spawner enrollment
	enrollment.InitializeEnrollmentManager()
	utils.PrintSection("Enrollment Manager", "ready", true)

	// Initialize SSE hub for real-time dashboard updates
	sseHub := sse.NewSSEHub()
	go sseHub.Run()
	utils.PrintSection("SSE Hub", "ready", true)

	// Initialize Router
	router := mux.NewRouter()

	// Initialize Database (PostgreSQL)
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		utils.PrintSection("Database", "disabled (no DB_DSN)", false)
	} else {
		var err error
		err = database.InitDB(dbDSN)
		if err != nil {
			utils.PrintSection("Database", "failed", false)
			utils.PrintSubItem(err.Error())
			registry.GlobalStats.SetDBConnected(false)
		} else {
			registry.GlobalStats.SetDBConnected(true)
			// Load existing spawners from DB to restore state
			loaded, _ := database.LoadSpawners(database.DBConn)
			if len(loaded) > 0 {
				maxID := registry.GetNextID() - 1
				for i := range loaded {
					s := loaded[i]
					copyS := s
					// Reconcile status: if marked online in DB but not actually connected via WS, set to offline
					if copyS.Status == "Online" && !ws.GlobalWSManager.IsClientConnected(copyS.ID) {
						log.Printf("INFO: Spawner ID %d was marked 'Online' in DB but not connected to WebSocket. Marking 'Offline'.", copyS.ID)
						copyS.Status = "Offline"
						// Optionally, persist this change to DB immediately, but WSManager will handle it on reconnect/disconnect
						// For now, only update in-memory to reflect reality for the Dashboard
					}
					registry.SetItem(copyS.ID, &copyS)
					if copyS.ID > maxID {
						maxID = copyS.ID
					}
				}
				utils.PrintSection("Database", "connected", true)
				utils.PrintSubItem(fmt.Sprintf("Loaded %d spawners", len(loaded)))
			} else {
				utils.PrintSection("Database", "connected", true)
			}

			// Initialize Player System Schema
			if err := database.InitPlayerSystem(database.DBConn); err != nil {
				log.Printf("Failed to init player system: %v", err)
			}
		}

		// Initialize Read-Only Database (Optional but recommended)
		roDSN := os.Getenv("READONLY_DB_DSN")
		if roDSN != "" {
			err = database.InitReadOnlyDB(roDSN)
			if err != nil {
				utils.PrintSection("Read-Only DB", "failed", false)
				utils.PrintSubItem(err.Error())

			} else {
				utils.PrintSection("Read-Only DB", "connected", true)
			}
		} else {
			// If not set, use main connection but fallback to regex validation
			database.ReadOnlyDBConn = database.DBConn
			if database.DBConn != nil {
				utils.PrintSection("Read-Only DB", "using primary (fallback)", true)
			}
		}
	}

	// Initialize RedEye Background Tasks
	if database.DBConn != nil {
		redeye.StartRedEyeBackground(database.DBConn)
	}

	// Initialize Firebase Remote Config
	_ = auth.InitFirebase()
	if auth.FirebaseMgr != nil && auth.FirebaseMgr.Connected {
		utils.PrintSection("Firebase", "connected", true)
		utils.PrintSubItem(fmt.Sprintf("Project: %s", auth.FirebaseMgr.ProjectID))
	} else {
		utils.PrintSection("Firebase", "disabled", false)
	}

	port := "8081"

	// Define API Routes

	// Spawner interactions (public/internal API) - Secured via API Key if set
	apiRouter := router.PathPrefix("/api/spawners").Subrouter()

	apiKey := os.Getenv("MASTER_API_KEY")
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"

	// Start WS Manager
	go ws.GlobalWSManager.Run()

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
	apiRouter.Use(middleware.UnifiedAuthMiddleware(apiKey, authConfig, sessionStore))

	apiRouter.HandleFunc("/ws", ws.GlobalWSManager.HandleWS) // WebSocket Endpoint
	apiRouter.HandleFunc("/download", handlers.ServeGameServerFile).Methods("GET", "HEAD")
	// apiRouter.HandleFunc("", RegisterSpawner).Methods("POST") // Disabled to enforce enrollment flow
	apiRouter.HandleFunc("", handlers.ListSpawners).Methods("GET") // Maybe this should be public or auth? Keeping consistent
	apiRouter.HandleFunc("/{id}", handlers.GetSpawner).Methods("GET")
	apiRouter.HandleFunc("/{id}", handlers.DeleteSpawner).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/spawn", handlers.SpawnInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/heartbeat", handlers.HeartbeatSpawner).Methods("POST")
	apiRouter.HandleFunc("/{id}/logs", handlers.GetSpawnerLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/logs", handlers.ClearSpawnerLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances", handlers.ListSpawnerInstances).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", handlers.GetInstanceLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", handlers.ClearInstanceLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats", handlers.GetInstanceStats).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/start", handlers.StartSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stop", handlers.StopSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restart", handlers.RestartSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/update", handlers.UpdateSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/rename", handlers.RenameSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}", handlers.RemoveSpawnerInstance).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup", handlers.BackupSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restore", handlers.RestoreSpawnerInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backups", handlers.ListSpawnerBackups).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup/delete", handlers.DeleteSpawnerBackup).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats/history", handlers.GetInstanceHistory).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/history", handlers.GetInstanceHistoryActions).Methods("GET")
	apiRouter.HandleFunc("/{id}/update-template", handlers.UpdateSpawnerTemplate).Methods("POST")

	// Liveness check
	router.HandleFunc("/health", handlers.Health).Methods("GET")

	// Authentication endpoints (API)
	authRouter := router.PathPrefix("/api/auth").Subrouter()

	authRouter.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLogin(w, r, authConfig, sessionStore)
	}).Methods("POST")

	authRouter.HandleFunc("/2fa", func(w http.ResponseWriter, r *http.Request) {
		auth.Handle2FAVerify(w, r, authConfig, sessionStore)
	}).Methods("POST")

	authRouter.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleEmailVerify(w, r, authConfig, sessionStore)
	}).Methods("POST")

	authRouter.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		auth.HandleLogout(w, r, sessionStore)
	}).Methods("GET", "POST")

	// Dashboard & UI endpoints (Protected by AuthMiddleware in dev/prod)
	statsHandler := http.HandlerFunc(handlers.StatsAPI)
	errorsAPIHandler := http.HandlerFunc(handlers.ErrorsAPI) // New API handler
	sseHandler := http.HandlerFunc(sseHub.HandleSSE)

	if authConfig.Enabled {
		router.Handle("/api/stats", auth.AuthMiddleware(authConfig, sessionStore)(statsHandler))
		router.Handle("/api/errors", auth.AuthMiddleware(authConfig, sessionStore)(errorsAPIHandler)).Methods("GET")
		router.Handle("/api/errors", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ClearErrorsAPI))).Methods("DELETE")
		router.Handle("/events", auth.AuthMiddleware(authConfig, sessionStore)(sseHandler)) // Replaced /ws with /events
		router.Handle("/api/upload", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.HandleUploadGameServer))).Methods("POST")

		// Version Management Routes
		router.Handle("/api/versions", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListVersions))).Methods("GET")
		router.Handle("/api/versions/{id}/active", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.HandleSetActiveVersion))).Methods("POST")
		router.Handle("/api/versions/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.HandleDeleteVersion))).Methods("DELETE")

		// Configuration Management Routes
		router.Handle("/api/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.GetAllConfigHandler))).Methods("GET")
		router.Handle("/api/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.CreateConfigHandler))).Methods("POST")
		router.Handle("/api/config/{category}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.GetConfigByCategoryHandler))).Methods("GET")
		router.Handle("/api/config/{key}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.UpdateConfigHandler))).Methods("PUT")
		router.Handle("/api/config/key/{key}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.GetConfigByKeyHandler))).Methods("GET")

		// Database Management Routes
		router.Handle("/api/database/tables", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetAllTablesHandler))).Methods("GET")
		router.Handle("/api/database/counts", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DebugListAllTablesHandler))).Methods("GET")
		router.Handle("/api/database/backup", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateInternalBackupHandler))).Methods("POST")

		router.Handle("/api/database/overview", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetDatabaseOverviewHandler))).Methods("GET")
		router.Handle("/api/database/table/{table}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetTableDataHandler))).Methods("GET")
		router.Handle("/api/database/table/{table}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.InsertTableRowHandler))).Methods("POST")
		router.Handle("/api/database/table/{table}/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.UpdateTableRowHandler))).Methods("PUT")
		router.Handle("/api/database/table/{table}/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DeleteTableRowHandler))).Methods("DELETE")
		router.Handle("/api/database/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetPostgresConfigHandler))).Methods("GET")
		router.Handle("/api/database/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.UpdatePostgresConfigHandler))).Methods("PUT")
		router.Handle("/api/database/config/restart", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.RestartPostgresHandler))).Methods("POST")

		// Introspection & SQL
		router.Handle("/api/database/schemas", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListSchemasHandler))).Methods("GET")
		router.Handle("/api/database/schemas", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateSchemaHandler))).Methods("POST")
		router.Handle("/api/database/schemas/{name}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DeleteSchemaHandler))).Methods("DELETE")
		router.Handle("/api/database/all-tables", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetAllTablesHandler))).Methods("GET")
		router.Handle("/api/database/tables", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListTablesBySchemaHandler))).Methods("GET")
		router.Handle("/api/database/tables/create", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateTableHandler))).Methods("POST")
		router.Handle("/api/database/tables/{table}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DropTableHandler))).Methods("DELETE")
		router.Handle("/api/database/tables/{table}/alter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.AlterTableHandler))).Methods("POST")
		router.Handle("/api/database/columns", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListColumnsHandler))).Methods("GET")
		router.Handle("/api/database/sql", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ExecuteSQLHandler))).Methods("POST")
		router.Handle("/api/database/debug/tables", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DebugListAllTablesHandler))).Methods("GET")

		// Roles
		router.Handle("/api/database/roles", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListRolesHandler))).Methods("GET")
		router.Handle("/api/database/roles", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateRoleHandler))).Methods("POST")
		router.Handle("/api/database/roles/{name}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DeleteRoleHandler))).Methods("DELETE")

		// Functions
		router.Handle("/api/database/functions", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListFunctionsHandler))).Methods("GET")
		router.Handle("/api/database/functions/details", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.GetFunctionHandler))).Methods("GET")
		router.Handle("/api/database/functions", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateFunctionHandler))).Methods("POST")
		router.Handle("/api/database/functions", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.UpdateFunctionHandler))).Methods("PUT")
		router.Handle("/api/database/functions/delete", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DeleteFunctionHandler))).Methods("POST")
		router.Handle("/api/database/functions/execute", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ExecuteFunctionHandler))).Methods("POST")

		router.Handle("/api/database/backups", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.CreateInternalBackupHandler))).Methods("POST")
		router.Handle("/api/database/backups", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.ListInternalBackupsHandler))).Methods("GET")
		router.Handle("/api/database/backups/{filename}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DownloadInternalBackupHandler))).Methods("GET")
		router.Handle("/api/database/backups/{filename}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.DeleteInternalBackupHandler))).Methods("DELETE")
		router.Handle("/api/database/restore", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(database.RestoreInternalBackupHandler))).Methods("POST")

		// System Management
		router.Handle("/api/restart", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.RestartServerHandler))).Methods("POST")

		// Performance Metrics API
		router.Handle("/api/metrics", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetAllMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/master", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetMasterMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/spawners", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetSpawnerMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/database", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetDatabaseMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/network", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetNetworkMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/gc", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ForceGCHandler))).Methods("POST")
		router.Handle("/api/metrics/memory/free", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.FreeMemoryHandler))).Methods("POST")

		// Enrollment Key Management (Dashboard authenticated endpoints)
		router.Handle("/api/enrollment/generate", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.GenerateEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/keys", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.ListEnrollmentKeysHandler))).Methods("GET")
		router.Handle("/api/enrollment/revoke", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.RevokeEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/status", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.GetEnrollmentKeyStatusHandler))).Methods("POST")

		// Firebase Remote Config Routes
		router.Handle("/api/config/firebase/status", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.GetFirebaseStatusHandler))).Methods("GET")
		router.Handle("/api/config/firebase/configs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.GetFirebaseConfigsHandler))).Methods("GET")
		router.Handle("/api/config/firebase/sync", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.SyncFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.CreateFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.UpdateFirebaseConfigHandler))).Methods("PUT")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.DeleteFirebaseConfigHandler))).Methods("DELETE")
		router.Handle("/api/config/firebase/publish", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.PublishFirebaseConfigHandler))).Methods("POST")

		// Notes & Todos API
		router.Handle("/api/notes", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListNotesHandler))).Methods("GET")
		router.Handle("/api/notes", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.CreateNoteHandler))).Methods("POST")
		router.Handle("/api/notes/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.UpdateNoteHandler))).Methods("PUT")
		router.Handle("/api/notes/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.DeleteNoteHandler))).Methods("DELETE")

		router.Handle("/api/todos", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListTodosHandler))).Methods("GET")
		router.Handle("/api/todos", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.CreateTodoHandler))).Methods("POST")
		router.Handle("/api/todos/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.UpdateTodoHandler))).Methods("PUT")
		router.Handle("/api/todos/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.DeleteTodoHandler))).Methods("DELETE")
		router.Handle("/api/todos/{id}/comments", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.CreateTodoCommentHandler))).Methods("POST")
		router.Handle("/api/todos/comments/{comment_id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.DeleteTodoCommentHandler))).Methods("DELETE")

		// Logging
		router.Handle("/api/logs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListSystemLogsHandler))).Methods("GET")
		router.Handle("/api/logs/counts", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetSystemLogCountsHandler))).Methods("GET")
		router.Handle("/api/logs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ClearSystemLogsHandler))).Methods("DELETE")
		router.Handle("/api/logs/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.DeleteSystemLogHandler))).Methods("DELETE")

		// RedEye Security System
		router.Handle("/api/redeye/stats", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.GetRedEyeStatsHandler))).Methods("GET")
		router.Handle("/api/redeye/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.GetRedEyeConfigHandler))).Methods("GET")
		router.Handle("/api/redeye/config", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.UpdateRedEyeConfigHandler))).Methods("PUT")
		router.Handle("/api/redeye/rules", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.ListRedEyeRulesHandler))).Methods("GET")
		router.Handle("/api/redeye/rules", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.CreateRedEyeRuleHandler))).Methods("POST")
		router.Handle("/api/redeye/rules/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.UpdateRedEyeRuleHandler))).Methods("PUT")
		router.Handle("/api/redeye/rules/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.DeleteRedEyeRuleHandler))).Methods("DELETE")
		router.Handle("/api/redeye/logs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.ListRedEyeLogsHandler))).Methods("GET")
		router.Handle("/api/redeye/logs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.ClearRedEyeLogsHandler))).Methods("DELETE")

		router.Handle("/api/redeye/bans", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.ListBannedIPsHandler))).Methods("GET")
		router.Handle("/api/redeye/bans/{ip}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.UnbanIPHandler))).Methods("DELETE")

		// RedEye Anti-Cheat
		router.Handle("/api/redeye/anticheat/report", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.ReportAnticheatEventHandler))).Methods("POST")
		router.Handle("/api/redeye/anticheat/events", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(redeye.GetAnticheatEventsHandler))).Methods("GET")

		// AI Bot API
		router.Handle("/api/ai/chat", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.AIChatHandler))).Methods("POST")

		// Game Player System
		router.Handle("/api/game/auth", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.AuthenticatePlayerHandler))).Methods("POST")
		router.Handle("/api/game/players", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListAllPlayersHandler))).Methods("GET")
		router.Handle("/api/game/players", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.CreateOrGetPlayerHandler))).Methods("POST")
		router.Handle("/api/game/players/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetPlayerDetailsHandler))).Methods("GET")
		
		router.Handle("/api/game/friends/request", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.SendFriendRequestHandler))).Methods("POST")
		router.Handle("/api/game/friends/accept", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.AcceptFriendRequestHandler))).Methods("POST")
	}

	// Enrollment endpoints (public - enrollment key IS the auth)
	router.HandleFunc("/api/enrollment/register", enrollment.RegisterSpawnerWithKeyHandler).Methods("POST")
	router.HandleFunc("/api/enrollment/validate", enrollment.ValidateEnrollmentKeyHandler).Methods("POST")

	// CLI-friendly status endpoint
	router.HandleFunc("/status", utils.PrintStatus).Methods("GET")

	// Start proactive health checks
	// go ProactiveHealthCheck(healthCheckInterval)

	// Start Stats Ticker (Memory & DB)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			registry.GlobalStats.UpdateMemoryStats()
			if database.DBConn != nil {
				registry.GlobalStats.UpdateDBStats(database.DBConn.Stats())
				// Advanced Stats
				advStats, err := database.GetAdvancedDBStats(database.DBConn)
				if err == nil {
					registry.GlobalStats.UpdateAdvancedDBStats(advStats)
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
		Handler: middleware.SecurityHeadersMiddleware(redeye.RedEyeMiddleware(middleware.StatsMiddleware(router))),
	}

	go func() {
		utils.PrintStartupComplete(port)
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
	if database.DBConn != nil {
		database.DBConn.Close()
	}
	log.Println("server stopped")
	return nil
}
