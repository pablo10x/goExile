package handlers

import (
	"net/http"
	"runtime"
	"runtime/debug"

	"exile/server/metrics"
	"exile/server/utils"
)

// GetAllMetricsHandler returns all combined metrics
func GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectAllMetrics()
	utils.WriteJSON(w, http.StatusOK, allMetrics)
}

// GetMasterMetricsHandler returns master server runtime metrics
func GetMasterMetricsHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectRuntimeMetrics()
	utils.WriteJSON(w, http.StatusOK, allMetrics)
}

// GetSpawnerMetricsHandler returns aggregated spawner metrics
func GetSpawnerMetricsHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectSpawnerMetrics()
	utils.WriteJSON(w, http.StatusOK, allMetrics)
}

// GetDatabaseMetricsHandler returns database metrics
func GetDatabaseMetricsHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectDatabaseMetrics()
	utils.WriteJSON(w, http.StatusOK, allMetrics)
}

// GetNetworkMetricsHandler returns network metrics
func GetNetworkMetricsHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectNetworkMetrics()
	utils.WriteJSON(w, http.StatusOK, allMetrics)
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

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
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

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
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
	pressure := metrics.GlobalMetrics.GetMemoryPressure()

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

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"pressure": pressure,
		"status":   status,
	})
}

// GetHealthCheckHandler returns a comprehensive health check
func GetHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := metrics.GlobalMetrics.CollectAllMetrics()

	// Determine overall health status
	issues := []string{}
	status := "healthy"

	// Check memory pressure
	pressure := metrics.GlobalMetrics.GetMemoryPressure()
	if pressure >= 80 {
		issues = append(issues, "High memory pressure")
		status = "critical"
	} else if pressure >= 60 && status != "critical" {
		issues = append(issues, "Moderate memory pressure")
		status = "warning"
	}

	// Check goroutine count
	if allMetrics.Master.NumGoroutine > 10000 {
		issues = append(issues, "High goroutine count")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check goroutine growth
	if allMetrics.Master.GoroutineGrowth > 100 {
		issues = append(issues, "Rapid goroutine growth detected")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check GC overhead
	if allMetrics.Master.GCCPUFraction > 0.05 { // More than 5% CPU on GC
		issues = append(issues, "High GC overhead")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check database connection
	if !allMetrics.Database.Connected {
		issues = append(issues, "Database disconnected")
		status = "critical"
	}

	// Check network error rate
	if allMetrics.Network.ErrorRate > 10 { // More than 10% error rate
		issues = append(issues, "High error rate")
		if status != "critical" {
			status = "warning"
		}
	}

	// Check spawner health
	if allMetrics.Spawners.TotalSpawners > 0 && allMetrics.Spawners.OnlineSpawners == 0 {
		issues = append(issues, "All spawners offline")
		status = "critical"
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"issues":          issues,
		"memory_pressure": pressure,
		"goroutines":      allMetrics.Master.NumGoroutine,
		"gc_cpu_fraction": allMetrics.Master.GCCPUFraction,
		"db_connected":    allMetrics.Database.Connected,
		"error_rate":      allMetrics.Network.ErrorRate,
		"spawners_online": allMetrics.Spawners.OnlineSpawners,
		"spawners_total":  allMetrics.Spawners.TotalSpawners,
		"uptime_ms":       allMetrics.Master.Uptime,
	})
}
