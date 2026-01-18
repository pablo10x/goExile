// Package main provides a lightweight HTTP-based node registry.
//
// The registry acts as the central authority for tracking active Nodes.
// It exposes a REST API for:
// - Node registration and heartbeats
// - Dashboard monitoring (via WebSocket)
// - Spawning new game instances (proxying to Nodes)
package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"exile/server/auth"
	"exile/server/config"
	"exile/server/database"
	"exile/server/enrollment"
	"exile/server/handlers"
	"exile/server/middleware"
	"exile/server/redeye"
	"exile/server/registry"
	"exile/server/utils"
	"exile/server/ws"
	"exile/server/ws_player"
)

// GzipResponseWriter wraps http.ResponseWriter to provide gzip compression
type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w GzipResponseWriter) Flush() {
	if flusher, ok := w.Writer.(*gzip.Writer); ok {
		_ = flusher.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// GzipMiddleware handles gzip compression for responses
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip WebSocket and SSE
		if r.Header.Get("Upgrade") == "websocket" || r.URL.Path == "/events" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := GzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

// LoggerMiddleware logs basic info about incoming requests
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("INCOMING: %s %s from %s (%v)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// Configuration constants used by the registry.
const (
	// maxBodySize is the maximum size accepted for request bodies.
	maxBodySize = 1 << 20 // 1MB

	// serverTTL defines how long a node is considered alive since its
	// last heartbeat. Nodes older than this are removed by cleanup.
	serverTTL = 60 * time.Second

	// cleanupInterval is how frequently the cleanup loop runs.
	cleanupInterval = 30 * time.Second

	// healthCheckInterval is how frequently the master server pings nodes for health checks.
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
	// Initialize TUI for startup
	p := tea.NewProgram(newTUIModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	model := m.(tuiModel)
	if model.err != nil {
		return model.err
	}

	// If credentials were generated, print them nicely
	if len(model.GeneratedCreds) > 0 {
		fmt.Println("\n================================================================================")
		fmt.Println("  ⚠️  SECURITY NOTICE: NEW CREDENTIALS GENERATED  ⚠️")
		fmt.Println("================================================================================")
		if key, ok := model.GeneratedCreds["MASTER_API_KEY"]; ok {
			fmt.Printf("  MASTER_API_KEY:  %s\n", key)
		}
		if key, ok := model.GeneratedCreds["GAME_API_KEY"]; ok {
			fmt.Printf("  GAME_API_KEY:    %s\n", key)
		}
		if pass, ok := model.GeneratedCreds["ADMIN_PASSWORD"]; ok {
			fmt.Printf("  ADMIN_PASSWORD:  %s\n", pass)
		}
		fmt.Println("================================================================================")
		fmt.Println("  These have been saved to your .env file.")
		fmt.Println("================================================================================")
		fmt.Println()
		// Pause briefly so user notices
		time.Sleep(3 * time.Second)
	}

	// Use GlobalStartup initialized in TUI steps
	authConfig := GlobalStartup.AuthConfig
	sessionStore := GlobalStartup.SessionStore
	sseHub := GlobalStartup.SSEHub
	router := GlobalStartup.Router

	// Create a root context for the application
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := "8081"

	// Define API Routes
	apiRouter := router.PathPrefix("/api/nodes").Subrouter()

	apiKey := os.Getenv("MASTER_API_KEY")
	gameAPIKey := os.Getenv("GAME_API_KEY")
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"

	if isProduction {
		if apiKey == "" {
			log.Fatal("FATAL: MASTER_API_KEY must be set in production mode")
		}
		if gameAPIKey == "" {
			log.Fatal("FATAL: GAME_API_KEY must be set in production mode")
		}
	} else {
		if apiKey == "" {
			apiKey = "dev_master_key" // Default for development
		}
		if gameAPIKey == "" {
			gameAPIKey = "dev_game_key" // Default for development
		}
	}

	// Always enforce Unified Auth (API Key OR Session)
	apiRouter.Use(middleware.UnifiedAuthMiddleware(apiKey, authConfig, sessionStore))

	apiRouter.HandleFunc("/ws", ws.GlobalWSManager.HandleWS) // WebSocket Endpoint
	apiRouter.HandleFunc("/download", handlers.ServeGameServerFile).Methods("GET", "HEAD")
	// apiRouter.HandleFunc("", RegisterNode).Methods("POST") // Disabled to enforce enrollment flow
	apiRouter.HandleFunc("", handlers.ListNodes).Methods("GET") // Maybe this should be public or auth? Keeping consistent
	apiRouter.HandleFunc("/{id}", handlers.GetNode).Methods("GET")
	apiRouter.HandleFunc("/{id}", handlers.UpdateNodeSettings).Methods("PUT") // New settings update endpoint
	apiRouter.HandleFunc("/{id}", handlers.DeleteNode).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/spawn", handlers.SpawnNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/heartbeat", handlers.HeartbeatNode).Methods("POST")
	apiRouter.HandleFunc("/{id}/logs", handlers.GetNodeLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/logs", handlers.ClearNodeLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances", handlers.ListNodeInstances).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", handlers.GetInstanceLogs).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/logs", handlers.ClearInstanceLogs).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats", handlers.GetInstanceStats).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/start", handlers.StartNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stop", handlers.StopNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restart", handlers.RestartNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/update", handlers.UpdateNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/rename", handlers.RenameNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}", handlers.RemoveNodeInstance).Methods("DELETE")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup", handlers.BackupNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/restore", handlers.RestoreNodeInstance).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backups", handlers.ListNodeBackups).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/backup/delete", handlers.DeleteNodeBackup).Methods("POST")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/stats/history", handlers.GetInstanceHistory).Methods("GET")
	apiRouter.HandleFunc("/{id}/instances/{instance_id:.+}/history", handlers.GetInstanceHistoryActions).Methods("GET")
	apiRouter.HandleFunc("/{id}/update-template", handlers.UpdateNodeTemplate).Methods("POST")

	// Game client routes - Secured via Game API Key
	gameRouter := router.PathPrefix("/api/game").Subrouter()
	gameRouter.Use(middleware.Auth_GameMiddleware(gameAPIKey))

	gameRouter.Handle("/auth", http.HandlerFunc(handlers.AuthenticatePlayerHandler)).Methods("POST")
	gameRouter.Handle("/ws", http.HandlerFunc(ws_player.GlobalPlayerWS.HandleWS))
	gameRouter.Handle("/players", http.HandlerFunc(handlers.ListAllPlayersHandler)).Methods("GET")
	gameRouter.Handle("/players", http.HandlerFunc(handlers.CreateOrGetPlayerHandler)).Methods("POST")
	gameRouter.Handle("/players/{id}", http.HandlerFunc(handlers.GetPlayerDetailsHandler)).Methods("GET")
	gameRouter.Handle("/friends/request", http.HandlerFunc(handlers.SendFriendRequestHandler)).Methods("POST")
	gameRouter.Handle("/friends/accept", http.HandlerFunc(handlers.AcceptFriendRequestHandler)).Methods("POST")
	gameRouter.Handle("/reports", http.HandlerFunc(handlers.CreateReportHandler)).Methods("POST")

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
		router.Handle("/api/config/category/{category}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.GetConfigByCategoryHandler))).Methods("GET")

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
		router.Handle("/api/metrics/nodes", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetNodeMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/database", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetDatabaseMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/network", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetNetworkMetricsHandler))).Methods("GET")
		router.Handle("/api/metrics/gc", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ForceGCHandler))).Methods("POST")
		router.Handle("/api/metrics/memory/free", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.FreeMemoryHandler))).Methods("POST")

		// Enrollment Key Management (Dashboard authenticated endpoints)
		router.Handle("/api/enrollment/generate", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.GenerateEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/keys", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.ListEnrollmentKeysHandler))).Methods("GET")
		router.Handle("/api/enrollment/revoke", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.RevokeEnrollmentKeyHandler))).Methods("POST")
		router.Handle("/api/enrollment/status", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.GetEnrollmentKeyStatusHandler))).Methods("POST")
		router.Handle("/api/enrollment/approve", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(enrollment.ApproveEnrollmentHandler))).Methods("POST")

		// Firebase Remote Config Routes
		router.Handle("/api/config/firebase/status", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.GetFirebaseStatusHandler))).Methods("GET")
		router.Handle("/api/config/firebase/configs", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.GetFirebaseConfigsHandler))).Methods("GET")
		router.Handle("/api/config/firebase/sync", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.SyncFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.CreateFirebaseConfigHandler))).Methods("POST")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.UpdateFirebaseConfigHandler))).Methods("PUT")
		router.Handle("/api/config/firebase/parameter", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.DeleteFirebaseConfigHandler))).Methods("DELETE")
		router.Handle("/api/config/firebase/publish", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(auth.PublishFirebaseConfigHandler))).Methods("POST")

		// General Config Routes (Moved to end to allow dots in keys)
		router.Handle("/api/config/{key:.+}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.UpdateConfigHandler))).Methods("PUT")
		router.Handle("/api/config/key/{key:.+}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(config.GetConfigByKeyHandler))).Methods("GET")

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

		// Fleet Management
		router.Handle("/api/instances", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListAllInstances))).Methods("GET")

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

		// Dashboard: Player Management (Session Protected)
		router.Handle("/api/admin/players", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListAllPlayersHandler))).Methods("GET")
		router.Handle("/api/admin/players/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.GetPlayerDetailsHandler))).Methods("GET")
		router.Handle("/api/admin/players/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.UpdatePlayerDetailsHandler))).Methods("PUT")
		router.Handle("/api/admin/players/{id}", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.DeletePlayerHandler))).Methods("DELETE")
		router.Handle("/api/admin/players/{id}/ban", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.BanPlayerHandler))).Methods("POST")

		// Dashboard: Reports (Session Protected)
		router.Handle("/api/reports", auth.AuthMiddleware(authConfig, sessionStore)(http.HandlerFunc(handlers.ListReportsHandler))).Methods("GET")
	}

	// Enrollment endpoints (public - enrollment key IS the auth)
	router.HandleFunc("/api/enrollment/register", enrollment.RegisterNodeWithKeyHandler).Methods("POST")
	router.HandleFunc("/api/enrollment/validate", enrollment.ValidateEnrollmentKeyHandler).Methods("POST")

	// CLI-friendly status endpoint
	router.HandleFunc("/status", utils.PrintStatus).Methods("GET")

	// Start proactive health checks
	// go ProactiveHealthCheck(healthCheckInterval)

	// Start Stats Ticker (Memory & DB)
	statsCtx, statsCancel := context.WithCancel(context.Background())
	defer statsCancel()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				registry.GlobalStats.UpdateMemoryStats()
				if database.DBConn != nil {
					registry.GlobalStats.UpdateDBStats(database.DBConn.Stats())
					// Advanced Stats
					advStats, err := database.GetAdvancedDBStats(database.DBConn)
					if err == nil {
						registry.GlobalStats.UpdateAdvancedDBStats(advStats)
					}
				}
			case <-statsCtx.Done():
				return
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
		Addr:         fmt.Sprintf("%s:%s", serverHost, port),
		Handler:      middleware.SecurityHeadersMiddleware(GzipMiddleware(redeye.RedEyeMiddleware(middleware.GlobalRateLimitMiddleware(middleware.StatsMiddleware(router))))),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		utils.PrintStartupComplete(port)
		// Warn if not binding to localhost
		if serverHost != "127.0.0.1" && serverHost != "localhost" {
			// Log a note about public binding if needed
		}

		// Wait for termination signal
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 10. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	log.Printf("received signal: %v", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop RedEye Engine
	redeye.StopRedEye()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
	if database.DBConn != nil {
		database.DBConn.Close()
	}
	log.Println("server stopped")
	return nil
}
