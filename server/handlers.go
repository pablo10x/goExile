package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	loginAttempts   = make(map[string]time.Time)
	loginAttemptsMu sync.Mutex
)

// ... (existing code remains until HandleLogin) ...

// RegisterSpawner accepts registration from a Spawner service.
func RegisterSpawner(w http.ResponseWriter, r *http.Request) {
	var s Spawner
	if err := decodeJSON(r, &s); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if s.Host == "" {
		writeError(w, r, http.StatusBadRequest, "invalid host")
		return
	}
	if s.Port < 1 || s.Port > 65535 {
		writeError(w, r, http.StatusBadRequest, "invalid port")
		return
	}
	if s.MaxInstances < 0 {
		writeError(w, r, http.StatusBadRequest, "invalid max_instances")
		return
	}

	id, err := registry.Register(&s)
	if err != nil {
		// Log the detailed error for debugging
		fmt.Printf("Registration failed: %v\n", err)
		writeError(w, r, http.StatusInternalServerError, "failed to register spawner")
		return
	}

	GlobalStats.UpdateActiveServers(len(registry.List()))
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatSpawner refreshes the LastSeen timestamp and stats for the given spawner ID.
func HeartbeatSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		CurrentInstances int     `json:"current_instances"`
		MaxInstances     int     `json:"max_instances"`
		Status           string  `json:"status"`
		CpuUsage         float64 `json:"cpu_usage"`
		MemUsed          uint64  `json:"mem_used"`
		MemTotal         uint64  `json:"mem_total"`
		DiskUsed         uint64  `json:"disk_used"`
		DiskTotal        uint64  `json:"disk_total"`
		GameVersion      string  `json:"game_version"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := registry.UpdateHeartbeat(id, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion); err != nil {
		writeError(w, r, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListSpawners returns a JSON array with all currently registered spawners.
func ListSpawners(w http.ResponseWriter, r *http.Request) {
	spawners := registry.List()
	GlobalStats.UpdateActiveServers(len(spawners))
	writeJSON(w, http.StatusOK, spawners)
}

// GetSpawner returns a single spawner by numeric id.
func GetSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// DeleteSpawner removes a spawner from the registry.
func DeleteSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.Delete(id) {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	GlobalStats.UpdateActiveServers(len(registry.List()))
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// SpawnInstance triggers a new game instance on the specified spawner.
func SpawnInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	id, err := parseID(rawID)
	if err != nil {
		fmt.Printf("Error parsing ID. Raw: '%s', Error: %v\n", rawID, err)
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "spawn", nil, 30*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		var resData struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(resp.Data, &resData) == nil && resData.ID != "" {
			SaveInstanceAction(dbConn, &InstanceAction{
				SpawnerID:  id,
				InstanceID: resData.ID,
				Action:     "spawn",
				Timestamp:  time.Now().UTC(),
				Status:     "success",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp.Data)
}

// GetSpawnerLogs fetches and returns the log file content from a spawner.
func GetSpawnerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "get_logs", nil, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ClearSpawnerLogs truncates the log file on a spawner.
func ClearSpawnerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "clear_logs", nil, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ListSpawnerInstances fetches active game instances from a spawner.
func ListSpawnerInstances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "list_instances", nil, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// UpdateSpawnerTemplate triggers the spawner to re-download the latest game server files from master.
func UpdateSpawnerTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "update_template", nil, 300*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// UpdateSpawnerInstance triggers an update (reinstall files) for a specific game instance.
func UpdateSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "update_instance", map[string]string{"instance_id": instanceID}, 300*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "update",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// RenameSpawnerInstance renames a specific game instance on a spawner.
func RenameSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		NewID string `json:"new_id"`
	}
	if err := decodeJSON(r, &reqBody); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "rename_instance", map[string]string{"instance_id": instanceID, "new_id": reqBody.NewID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "rename",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// RemoveSpawnerInstance removes a specific game instance on a spawner.
func RemoveSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "remove_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "delete",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// StopSpawnerInstance stops a specific game instance on a spawner.
func StopSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "stop_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "stop",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "instance stopped", "id": instanceID})
}

// StartSpawnerInstance starts a specific game instance on a spawner.
func StartSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	log.Printf("StartSpawnerInstance handler invoked for request: %s", r.URL.Path)
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "start_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "start",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "instance started", "id": instanceID})
}

// RestartSpawnerInstance restarts a specific game instance on a spawner.
func RestartSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "restart_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if dbConn != nil {
		SaveInstanceAction(dbConn, &InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "restart",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceLogs proxies the log stream from a specific game instance.
// Note: Returns full log content via WebSocket. For streaming, consider SSE or WebSocket streaming.
func GetInstanceLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "get_instance_logs", map[string]string{"instance_id": instanceID}, 30*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	// Return logs as JSON (for now - streaming can be added later via SSE or WebSocket)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ClearInstanceLogs proxies the request to clear an instance's logs.
func ClearInstanceLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "clear_instance_logs", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceStats proxies the stats request for a specific game instance.
func GetInstanceStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "get_instance_stats", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceHistory proxies the history request for a specific game instance.
func GetInstanceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "get_instance_history", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceHistoryActions retrieves the recorded action history for an instance.
func GetInstanceHistoryActions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	if dbConn == nil {
		writeJSON(w, http.StatusOK, []InstanceAction{})
		return
	}

	actions, err := GetInstanceActions(dbConn, id, instanceID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to retrieve history")
		return
	}
	writeJSON(w, http.StatusOK, actions)
}

// BackupSpawnerInstance creates a backup of a game instance on a spawner.
func BackupSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "backup_instance", map[string]string{"instance_id": instanceID}, 300*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// RestoreSpawnerInstance restores a backup of a game instance on a spawner.
func RestoreSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		Filename string `json:"filename"`
	}
	if err := decodeJSON(r, &reqBody); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "restore_instance", map[string]string{"instance_id": instanceID, "filename": reqBody.Filename}, 300*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ListSpawnerBackups lists backups of a game instance on a spawner.
func ListSpawnerBackups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "list_backups", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// DeleteSpawnerBackup deletes a backup of a game instance on a spawner.
func DeleteSpawnerBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		Filename string `json:"filename"`
	}
	if err := decodeJSON(r, &reqBody); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := GlobalWSManager.SendCommandSync(id, "delete_backup", map[string]string{"instance_id": instanceID, "filename": reqBody.Filename}, 10*time.Second)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		writeError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// Health is a lightweight liveness endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// StatsAPI returns JSON statistics about the server for the dashboard.
func StatsAPI(w http.ResponseWriter, r *http.Request) {
	totalReq, totalErr, active, dbOK, uptime, mem, tx, rx := GlobalStats.GetStats()

	stats := map[string]interface{}{
		"uptime":          uptime.Milliseconds(),
		"active_spawners": active,
		"total_requests":  totalReq,
		"total_errors":    totalErr,
		"db_connected":    dbOK,
		"memory_usage":    mem,
		"bytes_sent":      tx,
		"bytes_received":  rx,
	}

	writeJSON(w, http.StatusOK, stats)
}

// ErrorsAPI returns the recent error log.
func ErrorsAPI(w http.ResponseWriter, r *http.Request) {
	errors := GlobalStats.GetErrors()
	writeJSON(w, http.StatusOK, errors)
}

// ClearErrorsAPI clears the error log.
func ClearErrorsAPI(w http.ResponseWriter, r *http.Request) {
	GlobalStats.ClearErrors()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Error log cleared"})
}

// HandleLogin processes login requests.
func HandleLogin(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method == http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "GET login not supported. Please POST credentials.")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate Limiting: 1 attempt per 30 minutes per IP
	ip := r.RemoteAddr // In prod, consider X-Forwarded-For if behind proxy
	loginAttemptsMu.Lock()
	lastAttempt, exists := loginAttempts[ip]
	if exists && time.Since(lastAttempt) < 30*time.Minute {
		loginAttemptsMu.Unlock()
		// Silent failure: Redirect back to login without error message
		log.Printf("Login rate limited for IP: %s", ip)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	loginAttempts[ip] = time.Now()
	loginAttemptsMu.Unlock()

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Verify email and hashed password
	if email == authConfig.Email {
		err := bcrypt.CompareHashAndPassword([]byte(authConfig.HashedPassword), []byte(password))
		if err == nil { // Password matches
			sessionID, err := sessionStore.CreateSession()
			if err != nil {
				http.Error(w, "Failed to create session", http.StatusInternalServerError)
				return
			}

			// Secure attributes based on production mode
			sameSite := http.SameSiteLaxMode
			if authConfig.IsProduction {
				sameSite = http.SameSiteStrictMode
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   authConfig.IsProduction, // Only true in production
				SameSite: sameSite,
				MaxAge:   86400,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Authentication failed
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleLogout revokes the session.
func HandleLogout(w http.ResponseWriter, r *http.Request, sessionStore *SessionStore) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionStore.RevokeSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}