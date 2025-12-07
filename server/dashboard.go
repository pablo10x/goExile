package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ErrorLog represents a single error event.
type ErrorLog struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Message   string    `json:"message"`
	ClientIP  string    `json:"client_ip"`
}

// DashboardStats holds metrics for display.
type DashboardStats struct {
	mu              sync.RWMutex
	TotalRequests   int64
	TotalErrors     int64
	ActiveSpawners  int
	BytesSent       int64
	BytesReceived   int64
	MemUsage        uint64 // in bytes
	LastRequestTime time.Time
	DBConnected     bool
	Uptime          time.Duration
	StartTime       time.Time
	ErrorLogs       []ErrorLog
}

// GlobalStats is the global dashboard stats instance.
var GlobalStats = &DashboardStats{
	StartTime: time.Now(),
	ErrorLogs: make([]ErrorLog, 0),
}

// RecordRequest increments the request counter and bandwidth stats.
func (ds *DashboardStats) RecordRequest(statusCode int, bytesIn, bytesOut int64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.TotalRequests++
	ds.BytesReceived += bytesIn
	ds.BytesSent += bytesOut
	if statusCode >= 400 {
		ds.TotalErrors++
	}
	ds.LastRequestTime = time.Now()
}

// RecordError adds an error to the log.
func (ds *DashboardStats) RecordError(path string, status int, message string, clientIP string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
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

// UpdateMemoryStats updates the memory usage stat.
func (ds *DashboardStats) UpdateMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ds.mu.Lock()
	ds.MemUsage = m.Alloc
	ds.mu.Unlock()
}

// GetStats returns a snapshot of current stats.
func (ds *DashboardStats) GetStats() (totalReq int64, totalErr int64, active int, dbOK bool, uptime time.Duration, mem uint64, tx, rx int64) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
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

// GetErrors returns the list of recent errors.
func (ds *DashboardStats) GetErrors() []ErrorLog {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	// Return a copy to avoid races
	logs := make([]ErrorLog, len(ds.ErrorLogs))
	copy(logs, ds.ErrorLogs)
	return logs
}

// UpdateActiveServers updates the active server count.
func (ds *DashboardStats) UpdateActiveServers(count int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ActiveSpawners = count
}

// SetDBConnected updates DB connection status.
func (ds *DashboardStats) SetDBConnected(connected bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.DBConnected = connected
}

// PrintBanner prints a stylish ASCII banner.
func PrintBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════╗
║                 🎮 SPAWNER REGISTRY 🎮                        ║
║              xx             ║
╚═══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// PrintStatus prints real-time server status.
func PrintStatus(w http.ResponseWriter, r *http.Request) {
	// Simple record for status page itself
	GlobalStats.RecordRequest(http.StatusOK, 0, 0)

	totalReq, totalErr, active, dbOK, uptime, mem, tx, rx := GlobalStats.GetStats()

	dbStatus := "✓ Connected"
	if !dbOK {
		dbStatus = "✗ Disconnected"
	}

	status := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                    📊 REGISTRY STATUS 📊                       ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  🕐 Uptime:            %-42s║
║  🖥️  Active Spawners:   %-42d║
║  📡 Total Requests:    %-42d║
║  ⚠️  Errors:           %-42d║
║  💾 Memory Usage:      %-42s║
║  ⬆️  Tx:               %-42s║
║  ⬇️  Rx:               %-42s║
║  🗄️  Database:         %-42s║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`,
		formatDuration(uptime),
		active,
		totalReq,
		totalErr,
		formatBytes(mem),
		formatBytes(uint64(tx)),
		formatBytes(uint64(rx)),
		dbStatus,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, status)
}

// formatBytes formats bytes to human readable string
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration in a readable way.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// PrintSpawnerList prints formatted spawner list.
func PrintSpawnerList(spawners []*Spawner) string {
	if len(spawners) == 0 {
		return `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SPAWNERS 📦                            ║
╠════════════════════════════════════════════════════════════════╣
║  No spawners registered yet.                                   ║
╚════════════════════════════════════════════════════════════════╝
`
	}

	header := `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SPAWNERS 📦                            ║
╠════════════════════════════════════════════════════════════════╣
║ ID   │ Region           │ Host:Port         │ Status   │ Inst.  ║
╠════════════════════════════════════════════════════════════════╣
`

	body := ""
	for _, s := range spawners {
		status := "🟢 Online"
		if s.Status == "offline" {
			status = "🔴 Offline"
		}

		body += fmt.Sprintf("║ %-4d │ %-16s │ %-17s │ %-8s │ %d/%d  ║\n",
			s.ID,
			truncate(s.Region, 16),
			fmt.Sprintf("%s:%d", s.Host, s.Port),
			status,
			s.CurrentInstances,
			s.MaxInstances,
		)
	}

	footer := `╚════════════════════════════════════════════════════════════════╝
`

	return header + body + footer
}

// truncate truncates a string to max length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}

// PrintConfig prints server configuration.
func PrintConfig(port int, dbPath string) {
	config := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                   ⚙️  CONFIGURATION ⚙️                        ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  🌐 HTTP Port:         %-42d║
║  🗄️  Database Path:     %-42s║
║  🔧 Server TTL:        %-42ds║
║  🧹 Cleanup Interval:  %-42ds║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`,
		port,
		dbPath,
		serverTTL,
		cleanupInterval,
	)
	fmt.Println(config)
}