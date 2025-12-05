package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DashboardStats holds metrics for display.
type DashboardStats struct {
	mu              sync.RWMutex
	TotalRequests   int64
	TotalErrors     int64
	ActiveServers   int
	LastRequestTime time.Time
	DBConnected     bool
	Uptime          time.Duration
	StartTime       time.Time
}

// GlobalStats is the global dashboard stats instance.
var GlobalStats = &DashboardStats{
	StartTime: time.Now(),
}

// RecordRequest increments the request counter.
func (ds *DashboardStats) RecordRequest(statusCode int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.TotalRequests++
	if statusCode >= 400 {
		ds.TotalErrors++
	}
	ds.LastRequestTime = time.Now()
}

// GetStats returns a snapshot of current stats.
func (ds *DashboardStats) GetStats() (totalReq int64, totalErr int64, active int, dbOK bool, uptime time.Duration) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	totalReq = ds.TotalRequests
	totalErr = ds.TotalErrors
	active = ds.ActiveServers
	dbOK = ds.DBConnected
	uptime = time.Since(ds.StartTime)
	return
}

// UpdateActiveServers updates the active server count.
func (ds *DashboardStats) UpdateActiveServers(count int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ActiveServers = count
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
║                 🎮 GAME SERVER REGISTRY 🎮                    ║
║              Production-Ready Backend API v1.0               ║
╚═══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// PrintStatus prints real-time server status.
func PrintStatus(w http.ResponseWriter, r *http.Request) {
	GlobalStats.RecordRequest(http.StatusOK)

	totalReq, totalErr, active, dbOK, uptime := GlobalStats.GetStats()

	dbStatus := "✓ Connected"
	if !dbOK {
		dbStatus = "✗ Disconnected"
	}

	status := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                    📊 SERVER STATUS 📊                         ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  🕐 Uptime:            %-42s║
║  🖥️  Active Servers:    %-42d║
║  📡 Total Requests:    %-42d║
║  ⚠️  Errors:           %-42d║
║  🗄️  Database:         %-42s║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`,
		formatDuration(uptime),
		active,
		totalReq,
		totalErr,
		dbStatus,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, status)
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

// PrintServerList prints formatted server list.
func PrintServerList(servers []*Server) string {
	if len(servers) == 0 {
		return `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SERVERS 📦                             ║
╠════════════════════════════════════════════════════════════════╣
║  No servers registered yet.                                   ║
╚════════════════════════════════════════════════════════════════╝
`
	}

	header := `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SERVERS 📦                             ║
╠════════════════════════════════════════════════════════════════╣
║ ID   │ Name             │ Host:Port         │ Status   │ Players║
╠════════════════════════════════════════════════════════════════╣
`

	body := ""
	for _, s := range servers {
		status := "🟢 Online"
		if s.Status == "offline" {
			status = "🔴 Offline"
		}

		body += fmt.Sprintf("║ %-4d │ %-16s │ %-17s │ %-8s │ %d/%d  ║\n",
			s.ID,
			truncate(s.Name, 16),
			fmt.Sprintf("%s:%d", s.Host, s.Port),
			status,
			s.CurrentPlayers,
			s.MaxPlayers,
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
