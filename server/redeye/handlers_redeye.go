package redeye

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"

	"github.com/gorilla/mux"
)

// -- Stats Handler --

func GetRedEyeStatsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	stats, err := database.GetRedEyeStats(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Add dynamic system info
	response := map[string]interface{}{
		"total_rules":      stats.TotalRules,
		"active_bans":       stats.ActiveBans,
		"events_24h":        stats.Events24h,
		"logs_24h":          stats.Logs24h,
		"reputation_count":  stats.ReputationCount,
		"entropy":           stats.Entropy,
		"threat_level":      stats.ThreatLevel,
		"uptime":            stats.Uptime,
		"system_active":     RedEyeActive,
		"system_error":      RedEyeError,
		"node_id":           "MASTER_RE_01",
		"crc":               "0x8F2A11",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// -- Rules Handlers --

func ListRedEyeRulesHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	rules, err := database.GetRedEyeRules(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, rules)
}

func CreateRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var rule models.RedEyeRule
	if err := utils.DecodeJSON(r, &rule); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if rule.CIDR == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "CIDR is required")
		return
	}

	if rule.Action != "ALLOW" && rule.Action != "DENY" && rule.Action != "RATE_LIMIT" {
		utils.WriteError(w, r, http.StatusBadRequest, "Action must be ALLOW, DENY, or RATE_LIMIT")
		return
	}

	id, err := database.CreateRedEyeRule(database.DBConn, &rule)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	rule.ID = id
	utils.WriteJSON(w, http.StatusCreated, rule)
}

func UpdateRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid ID")
		return
	}

	var rule models.RedEyeRule
	if err := utils.DecodeJSON(r, &rule); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rule.ID = id

	if err := database.UpdateRedEyeRule(database.DBConn, &rule); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	utils.WriteJSON(w, http.StatusOK, rule)
}

func DeleteRedEyeRuleHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := database.DeleteRedEyeRule(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// -- Log Handlers --

func ListRedEyeLogsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	logs, total, err := database.GetRedEyeLogs(database.DBConn, limit, offset)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

func ClearRedEyeLogsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if err := database.ClearRedEyeLogs(database.DBConn); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

// -- Anticheat Handlers --

func ReportAnticheatEventHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var event models.RedEyeAnticheatEvent
	if err := utils.DecodeJSON(r, &event); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if event.ClientIP == "" {
		event.ClientIP = utils.GetClientIP(r)
	}
	event.Timestamp = time.Now()

	if err := database.SaveAnticheatEvent(database.DBConn, &event); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Logic for auto-banning based on reputation
	if event.Severity >= 5 {
		rep, _ := database.GetIPReputation(database.DBConn, event.ClientIP)
		if rep == nil {
			rep = &models.RedEyeIPReputation{
				IP:              event.ClientIP,
				ReputationScore: 0,
				IsBanned:        false,
				LastSeen:        time.Now(),
				TotalEvents:     0,
			}
		}

		rep.ReputationScore += event.Severity * 2
		rep.TotalEvents++
		rep.LastSeen = time.Now()

		if rep.ReputationScore >= 100 && !rep.IsBanned {
			rep.IsBanned = true
			rep.BanReason = fmt.Sprintf("Auto-banned due to high severity anticheat event: %s", event.EventType)
			// System call to block IP
			utils.BlockIPSystem(event.ClientIP)
			defer RefreshBanCache(database.DBConn)
		}

		database.UpdateIPReputation(database.DBConn, rep)
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "reported"})
}

func GetAnticheatEventsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	events, total, err := database.GetAnticheatEvents(database.DBConn, limit, offset)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
	})
}

func GetRedEyeConfigHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	config := map[string]interface{}{}
	keys := []string{"redeye.auto_ban_enabled", "redeye.auto_ban_threshold", "redeye.alert_enabled"}
	for _, key := range keys {
		// Note: we might need config package here, but for simplicity we assume database has GetConfigByKey
		// Wait, I should use database.GetConfigByKey
		cfg, err := database.GetConfigByKey(database.DBConn, key)
		if err == nil && cfg != nil {
			if key == "redeye.auto_ban_threshold" {
				val, _ := strconv.Atoi(cfg.Value)
				config[key] = val
			} else {
				config[key] = cfg.Value == "true"
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

func UpdateRedEyeConfigHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var payload map[string]interface{}
	if err := utils.DecodeJSON(r, &payload); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	allowedKeys := map[string]bool{
		"redeye.auto_ban_enabled":   true,
		"redeye.auto_ban_threshold": true,
		"redeye.alert_enabled":      true,
	}

	for k, v := range payload {
		if !allowedKeys[k] {
			continue
		}
		strVal := fmt.Sprintf("%v", v)
		if k == "redeye.auto_ban_threshold" {
			if f, ok := v.(float64); ok {
				strVal = fmt.Sprintf("%d", int(f))
			}
		}
		database.UpdateConfig(database.DBConn, k, strVal, "admin")
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func ListBannedIPsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	bans, err := database.GetBannedIPsFull(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, bans)
}

func UnbanIPHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	ip := vars["ip"]
	if ip == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "IP is required")
		return
	}

	if err := database.UnbanIP(database.DBConn, ip); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	RefreshBanCache(database.DBConn)
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}
