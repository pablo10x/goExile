package main

import (
	"net/http"
	"runtime"
	"runtime/debug"
)

// GetAllMetricsHandler returns all combined metrics
func GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectAllMetrics()
	writeJSON(w, http.StatusOK, metrics)
}

// GetMasterMetricsHandler returns master server runtime metrics
func GetMasterMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectRuntimeMetrics()
	writeJSON(w, http.StatusOK, metrics)
}

// GetSpawnerMetricsHandler returns aggregated spawner metrics
func GetSpawnerMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectSpawnerMetrics()
	writeJSON(w, http.StatusOK, metrics)
}

// GetDatabaseMetricsHandler returns database metrics
func GetDatabaseMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectDatabaseMetrics()
	writeJSON(w, http.StatusOK, metrics)
}

// GetNetworkMetricsHandler returns network metrics
func GetNetworkMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectNetworkMetrics()
	writeJSON(w, http.StatusOK, metrics)
}

// ForceGCHandler triggers a garbage collection cycle
func ForceGCHandler(w http.ResponseWriter, r *http.Request) {
	// Get stats before GC
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	heapBefore := mBefore.HeapAlloc
	numGCBefore := mBefore.NumGC

	// Run GC
	runtime.GC()

	// Get stats after GC
	var mAfter runtime.MemStats
	runtime.ReadMemStats(&mAfter)
	heapAfter := mAfter.HeapAlloc
	numGCAfter := mAfter.NumGC

	var freed uint64
	if heapBefore > heapAfter {
		freed = heapBefore - heapAfter
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Garbage collection completed",
		"heap_before":      heapBefore,
		"heap_after":       heapAfter,
		"freed_bytes":      freed,
		"gc_cycles_before": numGCBefore,
		"gc_cycles_after":  numGCAfter,
	})
}

// FreeMemoryHandler releases memory back to the OS
func FreeMemoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get stats before
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	sysBefore := mBefore.Sys
	heapReleasedBefore := mBefore.HeapReleased

	// Free OS memory
	debug.FreeOSMemory()

	// Get stats after
	var mAfter runtime.MemStats
	runtime.ReadMemStats(&mAfter)
	sysAfter := mAfter.Sys
	heapReleasedAfter := mAfter.HeapReleased

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":              "Memory released to OS",
		"sys_before":           sysBefore,
		"sys_after":            sysAfter,
		"heap_released_before": heapReleasedBefore,
		"heap_released_after":  heapReleasedAfter,
		"additional_released":  heapReleasedAfter - heapReleasedBefore,
	})
}

// GetMemoryPressureHandler returns the current memory pressure score
func GetMemoryPressureHandler(w http.ResponseWriter, r *http.Request) {
	pressure := GlobalMetrics.GetMemoryPressure()

	var status string
	switch {
	case pressure >= 80:
		status = "critical"
	case pressure >= 60:
		status = "warning"
	case pressure >= 40:
		status = "moderate"
	default:
		status = "healthy"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pressure": pressure,
		"status":   status,
	})
}

// GetHealthCheckHandler returns a comprehensive health check
func GetHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GlobalMetrics.CollectAllMetrics()

	// Determine overall health status
	issues := []string{}
	status := "healthy"

	// Check memory pressure
	pressure := GlobalMetrics.GetMemoryPressure()
	if pressure >= 80 {
		issues = append(issues, "High memory pressure")
		status = "critical"
	} else if pressure >= 60 && status != "critical" {
		issues = append(issues, "Moderate memory pressure")
		status = "warning"
	}

	// Check goroutine count
	if metrics.Master.NumGoroutine > 10000 {
		issues = append(issues, "High goroutine count")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check goroutine growth
	if metrics.Master.GoroutineGrowth > 100 {
		issues = append(issues, "Rapid goroutine growth detected")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check GC overhead
	if metrics.Master.GCCPUFraction > 0.05 { // More than 5% CPU on GC
		issues = append(issues, "High GC overhead")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check database connection
	if !metrics.Database.Connected {
		issues = append(issues, "Database disconnected")
		status = "critical"
	}

	// Check network error rate
	if metrics.Network.ErrorRate > 10 { // More than 10% error rate
		issues = append(issues, "High error rate")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check spawner health
	if metrics.Spawners.TotalSpawners > 0 && metrics.Spawners.OnlineSpawners == 0 {
		issues = append(issues, "All spawners offline")
		status = "critical"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"issues":          issues,
		"memory_pressure": pressure,
		"goroutines":      metrics.Master.NumGoroutine,
		"gc_cpu_fraction": metrics.Master.GCCPUFraction,
		"db_connected":    metrics.Database.Connected,
		"error_rate":      metrics.Network.ErrorRate,
		"spawners_online": metrics.Spawners.OnlineSpawners,
		"spawners_total":  metrics.Spawners.TotalSpawners,
		"uptime_ms":       metrics.Master.Uptime,
	})
}
