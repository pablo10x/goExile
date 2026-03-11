package security

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"

	"github.com/gorilla/mux"
)

// -- Stats Handler --

func GetSecurityStatsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	stats, err := database.GetSecurityStats(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	engineStats := GetEngineStats()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "MASTER_SEC_01"
	}

	// Generate a consistent CRC-like ID based on hostname
	crc := fmt.Sprintf("0x%X", utils.HashString(hostname)%0xFFFFFF)

	oldestLog, _ := database.GetOldestSecurityLogTimestamp(database.DBConn)

	// Add dynamic system info
	response := map[string]interface{}{
		"total_rules":      stats.TotalRules,
		"active_bans":      stats.ActiveBans, // DB persisted bans
		"events_24h":       stats.Events24h,
		"logs_24h":         stats.Logs24h,
		"reputation_count": stats.ReputationCount,
		"risk_factor":      stats.RiskFactor,
		"threat_level":     stats.ThreatLevel,
		"uptime":           stats.Uptime,
		"system_active":    SecurityActive,
		"system_error":     SecurityError,
		"node_id":          hostname,
		"crc":              crc,
		"oldest_log_at":    oldestLog,
		// Real-time Engine Metrics
		"rt_active_trackers": engineStats["active_trackers"],
		"rt_queue_depth":     engineStats["queue_depth"],
		"rt_cached_bans":     engineStats["cached_bans"],
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// -- Rules Handlers --

func ListSecurityRulesHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	rules, err := database.GetSecurityRules(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, rules)
}

func CreateSecurityRuleHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var rule models.SecurityRule
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

	if rule.Protocol == "" {
		rule.Protocol = "ANY"
	}

	ruleWithID, err := database.CreateSecurityRule(database.DBConn, &rule)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	utils.WriteJSON(w, http.StatusCreated, ruleWithID)
}

func UpdateSecurityRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	var rule models.SecurityRule
	if err := utils.DecodeJSON(r, &rule); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rule.ID = id

	if rule.Protocol == "" {
		rule.Protocol = "ANY"
	}

	if err := database.UpdateSecurityRule(database.DBConn, &rule); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	utils.WriteJSON(w, http.StatusOK, rule)
}

func DeleteSecurityRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := database.DeleteSecurityRule(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Refresh cache immediately
	RefreshBanCache(database.DBConn)

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// -- Log Handlers --

func ListSecurityLogsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	logs, total, err := database.GetSecurityLogs(database.DBConn, limit, offset)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

func ClearSecurityLogsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if err := database.ClearSecurityLogs(database.DBConn); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

// -- Security Handlers --

func ReportSecurityEventHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var event models.SecurityEvent
	if err := utils.DecodeJSON(r, &event); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if event.ClientIP == "" {
		event.ClientIP = utils.GetClientIP(r)
	}
	event.Timestamp = time.Now()

	// 1. Persist detailed evidence to permanent record
	if err := database.SaveSecurityEvent(database.DBConn, &event); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// 2. Feed the Real-time Engine
	details := fmt.Sprintf("[%s] Player: %s - %s", event.EventType, event.PlayerID, event.Details)
	IngestSignal(event.ClientIP, SignalTypeReport, event.Severity*2, details)

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "reported", "action": "processed_by_engine"})
}

func GetSecurityEventsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	events, total, err := database.GetSecurityEvents(database.DBConn, limit, offset)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
	})
}

func GetSecurityConfigHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	config := map[string]interface{}{}
	keys := []string{
		"security.auto_ban_enabled",
		"security.auto_ban_threshold",
		"security.alert_enabled",
		"security.mode",
		"security.strict_mode",
		"security.decay_rate",
		"security.whitelist_ips",
		"security.geoip_enabled",
		"security.allowed_countries",
	}
	for _, key := range keys {
		cfg, err := database.GetConfigByKey(database.DBConn, key)
		if err == nil && cfg != nil {
			if key == "security.auto_ban_threshold" || key == "security.decay_rate" {
				val, _ := strconv.Atoi(cfg.Value)
				config[key] = val
			} else if key == "security.auto_ban_enabled" || key == "security.alert_enabled" || key == "security.strict_mode" || key == "security.geoip_enabled" {
				config[key] = cfg.Value == "true"
			} else {
				config[key] = cfg.Value
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

func UpdateSecurityConfigHandler(w http.ResponseWriter, r *http.Request) {
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
		"security.auto_ban_enabled":   true,
		"security.auto_ban_threshold": true,
		"security.alert_enabled":      true,
		"security.mode":               true,
		"security.strict_mode":        true,
		"security.decay_rate":         true,
		"security.whitelist_ips":      true,
		"security.geoip_enabled":      true,
		"security.allowed_countries":  true,
	}

	for k, v := range payload {
		if !allowedKeys[k] {
			continue
		}
		strVal := fmt.Sprintf("%v", v)
		if k == "security.auto_ban_threshold" || k == "security.decay_rate" {
			if f, ok := v.(float64); ok {
				strVal = fmt.Sprintf("%d", int(f))
			}
		}
		database.UpdateConfig(database.DBConn, k, strVal, "admin")
	}

	syncConfig(database.DBConn)
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
