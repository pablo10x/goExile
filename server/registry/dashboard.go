package registry

import (
	"database/sql"
	"runtime"
	"sync"
	"time"

	"exile/server/database"
)

var (
	loginAttempts   = make(map[string]time.Time)
	loginAttemptsMu sync.Mutex
)

// ErrorLog represents a single error event.
type ErrorLog struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Message   string    `json:"message"`
	ClientIP  string    `json:"client_ip"`
}

// SecurityLog represents a security event.
type SecurityLog struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Details   string    `json:"details"`
	ClientIP  string    `json:"client_ip"`
}

// DashboardStats holds metrics for display.
type DashboardStats struct {
	Mu              sync.RWMutex
	TotalRequests   int64
	TotalErrors     int64
	ActiveSpawners  int
	BytesSent       int64
	BytesReceived   int64
	MemUsage        uint64
	LastRequestTime time.Time
	DBConnected     bool

	// DB Stats
	DBOpenConnections   int
	DBInUse             int
	DBIdle              int
	DBWaitCount         int64
	DBWaitDuration      time.Duration
	DBMaxLifetimeClosed int64
	DBMaxIdleClosed     int64

	// Advanced DB Stats
	DBSize        string
	DBCommits     int64
	DBRollbacks   int64
	DBCacheHit    float64
	DBTupFetched  int64
	DBTupInserted int64
	DBTupUpdated  int64
	DBTupDeleted  int64

	Uptime       time.Duration
	StartTime    time.Time
	ErrorLogs    []ErrorLog
	SecurityLogs []SecurityLog

	// RedEye Stats
	RedEyeTotalBlocks    int64
	RedEyeTotalRateLimit int64
	RedEyeActiveBans     int
}

// GlobalStats is the global dashboard stats instance.
var GlobalStats = &DashboardStats{
	StartTime:    time.Now(),
	ErrorLogs:    make([]ErrorLog, 0),
	SecurityLogs: make([]SecurityLog, 0),
}

// RecordRequest increments the request counter and bandwidth stats.
func (ds *DashboardStats) RecordRequest(statusCode int, bytesIn, bytesOut int64) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.TotalRequests++
	ds.BytesReceived += bytesIn
	ds.BytesSent += bytesOut
	// Errors are now recorded separately via RecordError with category check
	ds.LastRequestTime = time.Now()
}

// RecordError adds an error to the log.
func (ds *DashboardStats) RecordError(path string, status int, message string, clientIP string, category string) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()

	if category == "Internal" {
		ds.TotalErrors++
	}

	logEntry := ErrorLog{
		Timestamp: time.Now(),
		Path:      path,
		Status:    status,
		Message:   message,
		ClientIP:  clientIP,
	}

	// Prepend and keep last 50
	ds.ErrorLogs = append([]ErrorLog{logEntry}, ds.ErrorLogs...)
	if len(ds.ErrorLogs) > 50 {
		ds.ErrorLogs = ds.ErrorLogs[:50]
	}
}

// RecordSecurityEvent adds a security event to the log.
func (ds *DashboardStats) RecordSecurityEvent(event string, details string, clientIP string) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()

	logEntry := SecurityLog{
		Timestamp: time.Now(),
		Event:     event,
		Details:   details,
		ClientIP:  clientIP,
	}

	// Prepend and keep last 50
	ds.SecurityLogs = append([]SecurityLog{logEntry}, ds.SecurityLogs...)
	if len(ds.SecurityLogs) > 50 {
		ds.SecurityLogs = ds.SecurityLogs[:50]
	}
}

// RecordRedEyeBlock increments the RedEye block counter.
func (ds *DashboardStats) RecordRedEyeBlock() {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.RedEyeTotalBlocks++
}

// RecordRedEyeRateLimit increments the RedEye rate limit counter.
func (ds *DashboardStats) RecordRedEyeRateLimit() {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.RedEyeTotalRateLimit++
}

// UpdateRedEyeActiveBans updates the count of active bans.
func (ds *DashboardStats) UpdateRedEyeActiveBans(count int) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.RedEyeActiveBans = count
}

// UpdateMemoryStats updates the memory usage stat.
func (ds *DashboardStats) UpdateMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ds.Mu.Lock()
	ds.MemUsage = m.Alloc
	ds.Mu.Unlock()
}

// UpdateDBStats updates the database statistics.
func (ds *DashboardStats) UpdateDBStats(stats sql.DBStats) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.DBOpenConnections = stats.OpenConnections
	ds.DBInUse = stats.InUse
	ds.DBIdle = stats.Idle
	ds.DBWaitCount = stats.WaitCount
	ds.DBWaitDuration = stats.WaitDuration
	ds.DBMaxLifetimeClosed = stats.MaxLifetimeClosed
	ds.DBMaxIdleClosed = stats.MaxIdleClosed
}

// UpdateAdvancedDBStats updates advanced Postgres stats.
func (ds *DashboardStats) UpdateAdvancedDBStats(stats *database.AdvancedDBStats) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.DBSize = stats.DatabaseSize
	ds.DBCommits = stats.XactCommit
	ds.DBRollbacks = stats.XactRollback
	ds.DBCacheHit = stats.CacheHitRatio
	ds.DBTupFetched = stats.TupFetched
	ds.DBTupInserted = stats.TupInserted
	ds.DBTupUpdated = stats.TupUpdated
	ds.DBTupDeleted = stats.TupDeleted
}

// GetStats returns a snapshot of current stats.
func (ds *DashboardStats) GetStats() (totalReq int64, totalErr int64, active int, dbOK bool, uptime time.Duration, mem uint64, tx, rx int64) {
	ds.Mu.RLock()
	defer ds.Mu.RUnlock()
	totalReq = ds.TotalRequests
	totalErr = ds.TotalErrors
	active = ds.ActiveSpawners
	dbOK = ds.DBConnected
	uptime = time.Since(ds.StartTime)
	mem = ds.MemUsage
	tx = ds.BytesSent
	rx = ds.BytesReceived
	return
}

// GetStatsMap returns a map of all stats for JSON/SSE consumption.
func (ds *DashboardStats) GetStatsMap() map[string]interface{} {
	ds.Mu.RLock()
	defer ds.Mu.RUnlock()
	uptime := time.Since(ds.StartTime)
	return map[string]interface{}{
		"uptime":                 uptime.Milliseconds(),
		"active_spawners":        ds.ActiveSpawners,
		"total_requests":         ds.TotalRequests,
		"total_errors":           ds.TotalErrors,
		"db_connected":           ds.DBConnected,
		"memory_usage":           ds.MemUsage,
		"bytes_sent":             ds.BytesSent,
		"bytes_received":         ds.BytesReceived,
		"db_open_connections":    ds.DBOpenConnections,
		"db_in_use":              ds.DBInUse,
		"db_idle":                ds.DBIdle,
		"db_wait_count":          ds.DBWaitCount,
		"db_wait_duration":       ds.DBWaitDuration.String(),
		"db_max_lifetime_closed": ds.DBMaxLifetimeClosed,
		"db_max_idle_closed":     ds.DBMaxIdleClosed,
		"db_size":                ds.DBSize,
		"db_commits":             ds.DBCommits,
		"db_rollbacks":           ds.DBRollbacks,
		"db_cache_hit":           ds.DBCacheHit,
		"db_tup_fetched":         ds.DBTupFetched,
		"db_tup_inserted":        ds.DBTupInserted,
		"db_tup_updated":         ds.DBTupUpdated,
		"db_tup_deleted":         ds.DBTupDeleted,
	}
}

// GetErrors returns the list of recent errors.
func (ds *DashboardStats) GetErrors() []ErrorLog {
	ds.Mu.RLock()
	defer ds.Mu.RUnlock()
	// Return a copy to avoid races
	logs := make([]ErrorLog, len(ds.ErrorLogs))
	copy(logs, ds.ErrorLogs)
	return logs
}

// ClearErrors empties the error log and resets the error count.
func (ds *DashboardStats) ClearErrors() {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.ErrorLogs = make([]ErrorLog, 0)
	ds.TotalErrors = 0 // Reset the error counter
}

// UpdateActiveServers updates the active server count.
func (ds *DashboardStats) UpdateActiveServers(count int) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.ActiveSpawners = count
}

// SetDBConnected updates DB connection status.
func (ds *DashboardStats) SetDBConnected(connected bool) {
	ds.Mu.Lock()
	defer ds.Mu.Unlock()
	ds.DBConnected = connected
}

// ANSI color codes for terminal styling
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBgBlue  = "\033[44m"
)

// PrintBanner prints a stylish ASCII banner.
