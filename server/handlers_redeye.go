package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// -- Rules Handlers --

func ListRedEyeRulesHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	rules, err := GetRedEyeRules(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, rules)
}

func CreateRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var rule RedEyeRule
	if err := decodeJSON(r, &rule); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if rule.CIDR == "" {
		writeError(w, r, http.StatusBadRequest, "CIDR is required")
		return
	}
	if rule.Action != "ALLOW" && rule.Action != "DENY" && rule.Action != "RATE_LIMIT" {
		writeError(w, r, http.StatusBadRequest, "Action must be ALLOW, DENY, or RATE_LIMIT")
		return
	}

	id, err := CreateRedEyeRule(dbConn, &rule)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	rule.ID = id
	rule.CreatedAt = time.Now().UTC()
	writeJSON(w, http.StatusCreated, rule)
}

func UpdateRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid ID")
		return
	}

	var rule RedEyeRule
	if err := decodeJSON(r, &rule); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	rule.ID = id
	if err := UpdateRedEyeRule(dbConn, &rule); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, rule)
}

func DeleteRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := DeleteRedEyeRule(dbConn, id); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// -- Logs Handlers --

func ListRedEyeLogsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

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

	logs, total, err := GetRedEyeLogs(dbConn, limit, offset)
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

func ClearRedEyeLogsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if err := ClearRedEyeLogs(dbConn); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

// -- Anti-Cheat Handlers --

func ReportAnticheatEventHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var event RedEyeAnticheatEvent
	if err := decodeJSON(r, &event); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Enrich event
	event.Timestamp = time.Now().UTC()
	if event.ClientIP == "" {
		event.ClientIP = GetClientIP(r) // Default to reporter's IP if not provided
	}

	if err := SaveAnticheatEvent(dbConn, &event); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Update Reputation
	go func() {
		rep, _ := GetIPReputation(dbConn, event.ClientIP)
		if rep == nil {
			rep = &RedEyeIPReputation{
				IP:          event.ClientIP,
				TotalEvents: 0,
			}
		}
		rep.TotalEvents++
		rep.LastSeen = time.Now().UTC()
		// Simple severity scoring
		rep.ReputationScore += event.Severity / 10
		if rep.ReputationScore > 100 {
			rep.ReputationScore = 100
		}
		
		// Auto-ban logic
		if rep.ReputationScore >= 100 && !rep.IsBanned {
			rep.IsBanned = true
			rep.BanReason = "Auto-banned by RedEye: Reputation Score exceeded 100"
			// Create a DENY rule
			rule := &RedEyeRule{
				Name:      "Auto-Ban " + event.ClientIP,
				CIDR:      event.ClientIP,
				Port:      "*",
				Protocol:  "ANY",
				Action:    "DENY",
				Enabled:   true,
				CreatedAt: time.Now().UTC(),
			}
			CreateRedEyeRule(dbConn, rule)
		}

		UpdateIPReputation(dbConn, rep)
	}()

	writeJSON(w, http.StatusCreated, event)
}

func ListAnticheatEventsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	limit := 50
	offset := 0
	// ... (parsing logic similar to others)

	events, total, err := GetAnticheatEvents(dbConn, limit, offset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
	})
}