package metrics

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"exile/server/registry"
	"exile/server/ws"
)

// RuntimeMetrics holds comprehensive Go runtime statistics
type RuntimeMetrics struct {
	// Memory Stats
	HeapAlloc      uint64  `json:"heap_alloc"`       // Bytes allocated and in use
	HeapSys        uint64  `json:"heap_sys"`         // Bytes obtained from system
	HeapIdle       uint64  `json:"heap_idle"`        // Bytes in idle spans
	HeapInuse      uint64  `json:"heap_inuse"`       // Bytes in non-idle spans
	HeapReleased   uint64  `json:"heap_released"`    // Bytes released to OS
	HeapObjects    uint64  `json:"heap_objects"`     // Total allocated objects
	StackInuse     uint64  `json:"stack_inuse"`      // Bytes used by stack allocator
	StackSys       uint64  `json:"stack_sys"`        // Bytes obtained from system for stack
	MSpanInuse     uint64  `json:"mspan_inuse"`      // Bytes used by mspan structures
	MCacheInuse    uint64  `json:"mcache_inuse"`     // Bytes used by mcache structures
	BuckHashSys    uint64  `json:"buckhash_sys"`     // Bytes used by profiling bucket hash table
	GCSys          uint64  `json:"gc_sys"`           // Bytes used for garbage collection
	OtherSys       uint64  `json:"other_sys"`        // Other system allocations
	TotalAlloc     uint64  `json:"total_alloc"`      // Cumulative bytes allocated
	Sys            uint64  `json:"sys"`              // Total bytes obtained from system
	Mallocs        uint64  `json:"mallocs"`          // Cumulative count of heap allocations
	Frees          uint64  `json:"frees"`            // Cumulative count of heap frees
	LiveObjects    uint64  `json:"live_objects"`     // Mallocs - Frees
	HeapAllocRate  float64 `json:"heap_alloc_rate"`  // Bytes allocated per second
	HeapUsageRatio float64 `json:"heap_usage_ratio"` // HeapInuse / HeapSys

	// GC Stats
	NumGC          uint32  `json:"num_gc"`            // Number of completed GC cycles
	NumForcedGC    uint32  `json:"num_forced_gc"`     // Number of forced GC cycles
	GCCPUFraction  float64 `json:"gc_cpu_fraction"`   // Fraction of CPU used by GC
	LastGCPauseNs  uint64  `json:"last_gc_pause_ns"`  // Duration of last GC pause
	AvgGCPauseNs   uint64  `json:"avg_gc_pause_ns"`   // Average GC pause duration
	MaxGCPauseNs   uint64  `json:"max_gc_pause_ns"`   // Maximum GC pause duration
	TotalGCPauseNs uint64  `json:"total_gc_pause_ns"` // Total GC pause time
	LastGCTime     int64   `json:"last_gc_time"`      // Unix timestamp of last GC
	NextGCTarget   uint64  `json:"next_gc_target"`    // Target heap size for next GC
	GCTriggerRatio float64 `json:"gc_trigger_ratio"`  // Heap size growth trigger ratio

	// Goroutine Stats
	NumGoroutine    int   `json:"num_goroutine"`    // Current number of goroutines
	NumCPU          int   `json:"num_cpu"`          // Number of CPUs available
	NumCgoCall      int64 `json:"num_cgo_call"`     // Number of cgo calls
	GoroutineGrowth int   `json:"goroutine_growth"` // Goroutine count change since last check
	PeakGoroutines  int   `json:"peak_goroutines"`  // Peak goroutine count observed

	// Build Info
	GoVersion string `json:"go_version"` // Go runtime version
	GOOS      string `json:"goos"`       // Operating system
	GOARCH    string `json:"goarch"`     // Architecture

	// Timing
	Uptime          int64 `json:"uptime_ms"`          // Milliseconds since start
	CollectionTime  int64 `json:"collection_time_ns"` // Time to collect these metrics
	LastCollectedAt int64 `json:"last_collected_at"`  // Unix timestamp

	// History for trend analysis (last 60 samples, ~5 minutes at 5s interval)
	HeapAllocHistory   []uint64 `json:"heap_alloc_history,omitempty"`
	GoroutineHistory   []int    `json:"goroutine_history,omitempty"`
	GCPauseHistory     []uint64 `json:"gc_pause_history,omitempty"`
	RequestRateHistory []int64  `json:"request_rate_history,omitempty"`
}

// SpawnerMetrics holds aggregated metrics from all spawners
type SpawnerMetrics struct {
	TotalSpawners    int     `json:"total_spawners"`
	OnlineSpawners   int     `json:"online_spawners"`
	TotalInstances   int     `json:"total_instances"`
	RunningInstances int     `json:"running_instances"`
	TotalCPUUsage    float64 `json:"total_cpu_usage"`
	AvgCPUUsage      float64 `json:"avg_cpu_usage"`
	TotalMemUsed     uint64  `json:"total_mem_used"`
	TotalMemTotal    uint64  `json:"total_mem_total"`
	TotalDiskUsed    uint64  `json:"total_disk_used"`
	TotalDiskTotal   uint64  `json:"total_disk_total"`
	MemUsagePercent  float64 `json:"mem_usage_percent"`
	DiskUsagePercent float64 `json:"disk_usage_percent"`

	// Per-spawner breakdown
	SpawnerDetails []SpawnerDetail `json:"spawner_details,omitempty"`
}

// SpawnerDetail holds individual spawner metrics
type SpawnerDetail struct {
	ID               int     `json:"id"`
	Region           string  `json:"region"`
	Host             string  `json:"host"`
	Port             int     `json:"port"`
	Status           string  `json:"status"`
	CurrentInstances int     `json:"current_instances"`
	MaxInstances     int     `json:"max_instances"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemUsed          uint64  `json:"mem_used"`
	MemTotal         uint64  `json:"mem_total"`
	MemPercent       float64 `json:"mem_percent"`
	DiskUsed         uint64  `json:"disk_used"`
	DiskTotal        uint64  `json:"disk_total"`
	DiskPercent      float64 `json:"disk_percent"`
	GameVersion      string  `json:"game_version"`
	LastSeen         int64   `json:"last_seen"`
}

// CombinedMetrics holds all system metrics
type CombinedMetrics struct {
	Master   RuntimeMetrics  `json:"master"`
	Spawners SpawnerMetrics  `json:"spawners"`
	Database DatabaseMetrics `json:"database"`
	Network  NetworkMetrics  `json:"network"`
	RedEye   RedEyeMetrics   `json:"redeye"`
}

// RedEyeMetrics holds security-related metrics
type RedEyeMetrics struct {
	TotalBlocks          int64  `json:"total_blocks"`
	TotalRateLimits      int64  `json:"total_rate_limits"`
	ActiveBans           int    `json:"active_bans"`
	TotalRules           int    `json:"total_rules"`
	AvgProcessingTimeMs  float64 `json:"avg_processing_time_ms"`
	ThreatLevel          string `json:"threat_level"`
	LastBlockAt          string `json:"last_block_at"`
}

// DatabaseMetrics holds database-specific metrics
type DatabaseMetrics struct {
	Connected         bool    `json:"connected"`
	OpenConnections   int     `json:"open_connections"`
	InUse             int     `json:"in_use"`
	Idle              int     `json:"idle"`
	WaitCount         int64   `json:"wait_count"`
	WaitDurationMs    int64   `json:"wait_duration_ms"`
	MaxLifetimeClosed int64   `json:"max_lifetime_closed"`
	MaxIdleClosed     int64   `json:"max_idle_closed"`
	Size              string  `json:"size"`
	Commits           int64   `json:"commits"`
	Rollbacks         int64   `json:"rollbacks"`
	CacheHitRatio     float64 `json:"cache_hit_ratio"`
	TupFetched        int64   `json:"tup_fetched"`
	TupInserted       int64   `json:"tup_inserted"`
	TupUpdated        int64   `json:"tup_updated"`
	TupDeleted        int64   `json:"tup_deleted"`
}

// NetworkMetrics holds network-related statistics
type NetworkMetrics struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalErrors       int64   `json:"total_errors"`
	ErrorRate         float64 `json:"error_rate"`
	BytesSent         uint64  `json:"bytes_sent"`
	BytesReceived     uint64  `json:"bytes_received"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
	ActiveConnections int     `json:"active_connections"`

	// RedEye Stats
	RedEyeTotalBlocks    int64 `json:"redeye_total_blocks"`
	RedEyeTotalRateLimit int64 `json:"redeye_total_rate_limit"`
	RedEyeActiveBans     int   `json:"redeye_active_bans"`
}

// MetricsCollector manages metrics collection
type MetricsCollector struct {
	mu               sync.RWMutex
	startTime        time.Time
	lastCollectTime  time.Time
	lastHeapAlloc    uint64
	lastRequestCount int64
	peakGoroutines   int
	lastGoroutines   int

	// Historical data (circular buffers)
	heapHistory      []uint64
	goroutineHistory []int
	gcPauseHistory   []uint64
	requestHistory   []int64
	historyMaxLen    int
	historyIndex     int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:        time.Now(),
		lastCollectTime:  time.Now(),
		heapHistory:      make([]uint64, 0, 60),
		goroutineHistory: make([]int, 0, 60),
		gcPauseHistory:   make([]uint64, 0, 60),
		requestHistory:   make([]int64, 0, 60),
		historyMaxLen:    60, // 5 minutes at 5s intervals
	}
}

// CollectRuntimeMetrics gathers comprehensive Go runtime statistics
func (mc *MetricsCollector) CollectRuntimeMetrics() RuntimeMetrics {
	startTime := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Calculate rates
	timeDelta := time.Since(mc.lastCollectTime).Seconds()
	if timeDelta == 0 {
		timeDelta = 1
	}

	heapAllocRate := float64(m.TotalAlloc-mc.lastHeapAlloc) / timeDelta
	mc.lastHeapAlloc = m.TotalAlloc

	// Track goroutine changes
	currentGoroutines := runtime.NumGoroutine()
	goroutineGrowth := currentGoroutines - mc.lastGoroutines
	mc.lastGoroutines = currentGoroutines
	if currentGoroutines > mc.peakGoroutines {
		mc.peakGoroutines = currentGoroutines
	}

	// Calculate GC statistics
	var lastGCPause, avgGCPause, maxGCPause uint64
	var totalGCPause uint64 = m.PauseTotalNs

	if m.NumGC > 0 {
		// Get last GC pause
		lastPauseIdx := (m.NumGC + 255) % 256
		lastGCPause = m.PauseNs[lastPauseIdx]

		// Calculate average and find max
		var sum uint64
		count := uint32(256)
		if m.NumGC < 256 {
			count = m.NumGC
		}
		for i := uint32(0); i < count; i++ {
			pause := m.PauseNs[i]
			sum += pause
			if pause > maxGCPause {
				maxGCPause = pause
			}
		}
		avgGCPause = sum / uint64(count)
	}

	// Get heap usage ratio
	var heapUsageRatio float64
	if m.HeapSys > 0 {
		heapUsageRatio = float64(m.HeapInuse) / float64(m.HeapSys)
	}

	// Get GC target ratio from debug info
	var gcTriggerRatio float64
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		_ = buildInfo // Could extract more info if needed
	}
	// Default GOGC is 100, meaning trigger at 100% heap growth
	gcTriggerRatio = float64(debug.SetGCPercent(-1))
	debug.SetGCPercent(int(gcTriggerRatio))
	gcTriggerRatio /= 100.0

	// Update history
	mc.appendUint64History(&mc.heapHistory, m.HeapAlloc)
	mc.appendIntHistory(&mc.goroutineHistory, currentGoroutines)
	mc.appendUint64History(&mc.gcPauseHistory, lastGCPause)

	mc.lastCollectTime = time.Now()

	metrics := RuntimeMetrics{
		// Memory
		HeapAlloc:      m.HeapAlloc,
		HeapSys:        m.HeapSys,
		HeapIdle:       m.HeapIdle,
		HeapInuse:      m.HeapInuse,
		HeapReleased:   m.HeapReleased,
		HeapObjects:    m.HeapObjects,
		StackInuse:     m.StackInuse,
		StackSys:       m.StackSys,
		MSpanInuse:     m.MSpanInuse,
		MCacheInuse:    m.MCacheInuse,
		BuckHashSys:    m.BuckHashSys,
		GCSys:          m.GCSys,
		OtherSys:       m.OtherSys,
		TotalAlloc:     m.TotalAlloc,
		Sys:            m.Sys,
		Mallocs:        m.Mallocs,
		Frees:          m.Frees,
		LiveObjects:    m.Mallocs - m.Frees,
		HeapAllocRate:  heapAllocRate,
		HeapUsageRatio: heapUsageRatio,

		// GC
		NumGC:          m.NumGC,
		NumForcedGC:    m.NumForcedGC,
		GCCPUFraction:  m.GCCPUFraction,
		LastGCPauseNs:  lastGCPause,
		AvgGCPauseNs:   avgGCPause,
		MaxGCPauseNs:   maxGCPause,
		TotalGCPauseNs: totalGCPause,
		LastGCTime:     int64(m.LastGC),
		NextGCTarget:   m.NextGC,
		GCTriggerRatio: gcTriggerRatio,

		// Goroutines
		NumGoroutine:    currentGoroutines,
		NumCPU:          runtime.NumCPU(),
		NumCgoCall:      runtime.NumCgoCall(),
		GoroutineGrowth: goroutineGrowth,
		PeakGoroutines:  mc.peakGoroutines,

		// Build info
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,

		// Timing
		Uptime:          time.Since(mc.startTime).Milliseconds(),
		CollectionTime:  time.Since(startTime).Nanoseconds(),
		LastCollectedAt: time.Now().Unix(),

		// History
		HeapAllocHistory: mc.copyUint64History(mc.heapHistory),
		GoroutineHistory: mc.copyIntHistory(mc.goroutineHistory),
		GCPauseHistory:   mc.copyUint64History(mc.gcPauseHistory),
	}

	return metrics
}

// CollectSpawnerMetrics gathers metrics from all registered spawners
func (mc *MetricsCollector) CollectSpawnerMetrics() SpawnerMetrics {
	spawners := registry.GlobalRegistry.All()

	metrics := SpawnerMetrics{
		TotalSpawners:  len(spawners),
		SpawnerDetails: make([]SpawnerDetail, 0, len(spawners)),
	}

	for _, s := range spawners {
		// Calculate percentages
		var memPercent, diskPercent float64
		if s.MemTotal > 0 {
			memPercent = (float64(s.MemUsed) / float64(s.MemTotal)) * 100
		}
		if s.DiskTotal > 0 {
			diskPercent = (float64(s.DiskUsed) / float64(s.DiskTotal)) * 100
		}

		detail := SpawnerDetail{
			ID:               s.ID,
			Region:           s.Region,
			Host:             s.Host,
			Port:             s.Port,
			Status:           s.Status,
			CurrentInstances: s.CurrentInstances,
			MaxInstances:     s.MaxInstances,
			CPUUsage:         s.CpuUsage,
			MemUsed:          s.MemUsed,
			MemTotal:         s.MemTotal,
			MemPercent:       memPercent,
			DiskUsed:         s.DiskUsed,
			DiskTotal:        s.DiskTotal,
			DiskPercent:      diskPercent,
			GameVersion:      s.GameVersion,
			LastSeen:         s.LastSeen.Unix(),
		}

		metrics.SpawnerDetails = append(metrics.SpawnerDetails, detail)

		// Aggregate totals
		if s.Status == "Online" {
			metrics.OnlineSpawners++
		}
		metrics.TotalInstances += s.CurrentInstances
		if s.Status == "Online" {
			metrics.RunningInstances += s.CurrentInstances
		}
		metrics.TotalCPUUsage += s.CpuUsage
		metrics.TotalMemUsed += s.MemUsed
		metrics.TotalMemTotal += s.MemTotal
		metrics.TotalDiskUsed += s.DiskUsed
		metrics.TotalDiskTotal += s.DiskTotal
	}

	// Calculate averages
	if metrics.OnlineSpawners > 0 {
		metrics.AvgCPUUsage = metrics.TotalCPUUsage / float64(metrics.OnlineSpawners)
	}
	if metrics.TotalMemTotal > 0 {
		metrics.MemUsagePercent = (float64(metrics.TotalMemUsed) / float64(metrics.TotalMemTotal)) * 100
	}
	if metrics.TotalDiskTotal > 0 {
		metrics.DiskUsagePercent = (float64(metrics.TotalDiskUsed) / float64(metrics.TotalDiskTotal)) * 100
	}

	return metrics
}

// CollectDatabaseMetrics gathers database statistics
func (mc *MetricsCollector) CollectDatabaseMetrics() DatabaseMetrics {
	metrics := DatabaseMetrics{}

	registry.GlobalStats.Mu.RLock()
	defer registry.GlobalStats.Mu.RUnlock()

	metrics.Connected = registry.GlobalStats.DBConnected
	metrics.OpenConnections = registry.GlobalStats.DBOpenConnections
	metrics.InUse = registry.GlobalStats.DBInUse
	metrics.Idle = registry.GlobalStats.DBIdle
	metrics.WaitCount = registry.GlobalStats.DBWaitCount
	metrics.WaitDurationMs = registry.GlobalStats.DBWaitDuration.Milliseconds()
	metrics.MaxLifetimeClosed = registry.GlobalStats.DBMaxLifetimeClosed
	metrics.MaxIdleClosed = registry.GlobalStats.DBMaxIdleClosed
	metrics.Size = registry.GlobalStats.DBSize
	metrics.Commits = registry.GlobalStats.DBCommits
	metrics.Rollbacks = registry.GlobalStats.DBRollbacks
	metrics.CacheHitRatio = registry.GlobalStats.DBCacheHit
	metrics.TupFetched = registry.GlobalStats.DBTupFetched
	metrics.TupInserted = registry.GlobalStats.DBTupInserted
	metrics.TupUpdated = registry.GlobalStats.DBTupUpdated
	metrics.TupDeleted = registry.GlobalStats.DBTupDeleted

	return metrics
}

// CollectNetworkMetrics gathers network-related statistics
func (mc *MetricsCollector) CollectNetworkMetrics() NetworkMetrics {
	registry.GlobalStats.Mu.RLock()
	totalReqs := registry.GlobalStats.TotalRequests
	totalErrs := registry.GlobalStats.TotalErrors
	bytesSent := registry.GlobalStats.BytesSent
	bytesRecv := registry.GlobalStats.BytesReceived
	redeyeBlocks := registry.GlobalStats.RedEyeTotalBlocks
	redeyeRateLimit := registry.GlobalStats.RedEyeTotalRateLimit
	redeyeActiveBans := registry.GlobalStats.RedEyeActiveBans
	registry.GlobalStats.Mu.RUnlock()
	mc.mu.Lock()
	// Calculate request rate
	timeDelta := time.Since(mc.lastCollectTime).Seconds()
	if timeDelta == 0 {
		timeDelta = 1
	}
	requestRate := float64(totalReqs-mc.lastRequestCount) / timeDelta
	mc.lastRequestCount = totalReqs

	// Update request history
	mc.appendInt64History(&mc.requestHistory, totalReqs)
	mc.mu.Unlock()

	var errorRate float64
	if totalReqs > 0 {
		errorRate = (float64(totalErrs) / float64(totalReqs)) * 100
	}

	// Count active WebSocket connections
	activeConnections := 0
	ws.GlobalWSManager.Mu.RLock()
	activeConnections = len(ws.GlobalWSManager.Connections)
	ws.GlobalWSManager.Mu.RUnlock()

	return NetworkMetrics{

		TotalRequests: totalReqs,

		TotalErrors: totalErrs,

		ErrorRate: errorRate,

		BytesSent: uint64(bytesSent),

		BytesReceived: uint64(bytesRecv),

		RequestsPerSecond: requestRate,

		ActiveConnections: activeConnections,

		RedEyeTotalBlocks: redeyeBlocks,

		RedEyeTotalRateLimit: redeyeRateLimit,

		RedEyeActiveBans: redeyeActiveBans,
	}

}

// CollectAllMetrics gathers all metrics in one call
func (mc *MetricsCollector) CollectAllMetrics() CombinedMetrics {
	return CombinedMetrics{
		Master:   mc.CollectRuntimeMetrics(),
		Spawners: mc.CollectSpawnerMetrics(),
		Database: mc.CollectDatabaseMetrics(),
		Network:  mc.CollectNetworkMetrics(),
		RedEye:   mc.CollectRedEyeMetrics(),
	}
}

// CollectRedEyeMetrics gathers security statistics
func (mc *MetricsCollector) CollectRedEyeMetrics() RedEyeMetrics {
	registry.GlobalStats.Mu.RLock()
	defer registry.GlobalStats.Mu.RUnlock()

	threatLevel := "LOW"
	if registry.GlobalStats.RedEyeActiveBans > 100 || registry.GlobalStats.RedEyeTotalBlocks > 5000 {
		threatLevel = "CRITICAL"
	} else if registry.GlobalStats.RedEyeActiveBans > 20 || registry.GlobalStats.RedEyeTotalBlocks > 1000 {
		threatLevel = "HIGH"
	} else if registry.GlobalStats.RedEyeActiveBans > 5 || registry.GlobalStats.RedEyeTotalBlocks > 100 {
		threatLevel = "MODERATE"
	}

	return RedEyeMetrics{
		TotalBlocks:         registry.GlobalStats.RedEyeTotalBlocks,
		TotalRateLimits:     registry.GlobalStats.RedEyeTotalRateLimit,
		ActiveBans:          registry.GlobalStats.RedEyeActiveBans,
		TotalRules:           0, // Would need redeye registry for this
		AvgProcessingTimeMs: 0.12, // Placeholder for actual timing
		ThreatLevel:         threatLevel,
		LastBlockAt:         time.Now().Format(time.RFC3339), // Placeholder
	}
}

// Helper functions for history management

func (mc *MetricsCollector) appendUint64History(history *[]uint64, value uint64) {
	if len(*history) >= mc.historyMaxLen {
		// Shift left and append
		*history = append((*history)[1:], value)
	} else {
		*history = append(*history, value)
	}
}

func (mc *MetricsCollector) appendIntHistory(history *[]int, value int) {
	if len(*history) >= mc.historyMaxLen {
		// Shift left and append
		*history = append((*history)[1:], value)
	} else {
		*history = append(*history, value)
	}
}

func (mc *MetricsCollector) appendInt64History(history *[]int64, value int64) {
	if len(*history) >= mc.historyMaxLen {
		// Shift left and append
		*history = append((*history)[1:], value)
	} else {
		*history = append(*history, value)
	}
}

func (mc *MetricsCollector) copyUint64History(src []uint64) []uint64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]uint64, len(src))
	copy(dst, src)
	return dst
}

func (mc *MetricsCollector) copyIntHistory(src []int) []int {
	if len(src) == 0 {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

// ForceGC triggers a garbage collection (use sparingly)
func (mc *MetricsCollector) ForceGC() {
	runtime.GC()
}

// FreeOSMemory returns memory to the OS (use sparingly)
func (mc *MetricsCollector) FreeOSMemory() {
	debug.FreeOSMemory()
}

// GetMemoryPressure returns a 0-100 score indicating memory pressure
func (mc *MetricsCollector) GetMemoryPressure() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate pressure based on multiple factors
	pressure := 0.0

	// Heap usage ratio (0-40 points)
	if m.HeapSys > 0 {
		heapRatio := float64(m.HeapInuse) / float64(m.HeapSys)
		pressure += heapRatio * 40
	}

	// GC frequency indicator (0-30 points)
	// High GC CPU fraction indicates memory pressure
	pressure += m.GCCPUFraction * 3000 // Scale up since it's usually < 1%

	// Live objects growth (0-30 points)
	// More objects = more memory pressure
	if m.HeapObjects > 1000000 {
		pressure += 30
	} else if m.HeapObjects > 100000 {
		pressure += 15
	}

	if pressure > 100 {
		pressure = 100
	}

	return int(pressure)
}

// Global metrics collector instance
var GlobalMetrics = NewMetricsCollector()
