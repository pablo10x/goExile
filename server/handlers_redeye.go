package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// -- Stats Handler --

func GetRedEyeStatsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	stats, err := GetRedEyeStats(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

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
	// Trigger refresh for immediate effect
	RefreshBanCache(dbConn)
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

	RefreshBanCache(dbConn)
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

	RefreshBanCache(dbConn)
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
			// Refresh cache to enforce ban immediately
			defer RefreshBanCache(dbConn)
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

// -- Config Handlers --

func GetRedEyeConfigHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	config := map[string]interface{}{}
	
	keys := []string{"redeye.auto_ban_enabled", "redeye.auto_ban_threshold", "redeye.alert_enabled"}
	for _, key := range keys {
		cfg, err := GetConfigByKey(dbConn, key)
		if err == nil && cfg != nil {
			// Convert to appropriate type if needed, or send as string
			if key == "redeye.auto_ban_threshold" {
				val, _ := strconv.Atoi(cfg.Value)
				config[key] = val
			} else {
				config[key] = cfg.Value == "true"
			}
		}
	}

	writeJSON(w, http.StatusOK, config)
}

func UpdateRedEyeConfigHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var payload map[string]interface{}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Update specific keys only
	allowedKeys := map[string]bool{
		"redeye.auto_ban_enabled": true,
		"redeye.auto_ban_threshold": true,
		"redeye.alert_enabled": true,
	}

	for k, v := range payload {
		if !allowedKeys[k] {
			continue
		}
		
		strVal := fmt.Sprintf("%v", v)
		if k == "redeye.auto_ban_threshold" {
			// Verify int
			if _, ok := v.(float64); ok { // JSON numbers are float64
				strVal = fmt.Sprintf("%d", int(v.(float64)))
			}
		}

		// Use "system" or authenticated user as updater. 
		// Since we use AuthMiddleware, we could extract user, but "admin" is fine for now.
		if err := UpdateConfig(dbConn, k, strVal, "admin"); err != nil {
			writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update %s: %v", k, err))
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// -- Ban Management Handlers --

func ListBannedIPsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	bans, err := GetBannedIPsFull(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bans)
}

func UnbanIPHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	ip := vars["ip"]
	if ip == "" {
		writeError(w, r, http.StatusBadRequest, "IP is required")
		return
	}

	if err := UnbanIP(dbConn, ip); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(dbConn)

	writeJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}
