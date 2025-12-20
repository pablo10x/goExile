package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ListSystemLogsHandler returns system logs with filtering.
func ListSystemLogsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	category := r.URL.Query().Get("category")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, total, err := GetSystemLogs(dbConn, category, limit, offset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetSystemLogCountsHandler returns log counts by category.
func GetSystemLogCountsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	counts, err := GetSystemLogCounts(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, counts)
}

// DeleteSystemLogHandler deletes a single log entry.
func DeleteSystemLogHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid log id")
		return
	}

	if err := DeleteSystemLog(dbConn, id); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ClearSystemLogsHandler deletes all system logs.
func ClearSystemLogsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if err := ClearSystemLogs(dbConn); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}
