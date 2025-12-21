package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/registry"
	"exile/server/utils"
	"exile/server/ws"

	"github.com/gorilla/mux"
)

// ... (existing code remains until HandleLogin) ...

// RegisterSpawner accepts registration from a models.Spawner service.
func RegisterSpawner(w http.ResponseWriter, r *http.Request) {
	var s models.Spawner
	if err := utils.DecodeJSON(r, &s); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if s.Host == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid host")
		return
	}
	if s.Port < 1 || s.Port > 65535 {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid port")
		return
	}
	if s.MaxInstances < 0 {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid max_instances")
		return
	}

	var id int
	var err error
	if id, err = registry.GlobalRegistry.Register(&s); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to register")
		return
	}

	registry.GlobalStats.UpdateActiveServers(len(registry.GlobalRegistry.List()))
	utils.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatSpawner refreshes the LastSeen timestamp and stats for the given spawner ID.
func HeartbeatSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
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
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := registry.GlobalRegistry.UpdateHeartbeat(id, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion); err != nil {
		utils.WriteError(w, r, http.StatusNotFound, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListSpawners returns a JSON array with all currently registered spawners.
func ListSpawners(w http.ResponseWriter, r *http.Request) {
	spawners := registry.GlobalRegistry.List()
	registry.GlobalStats.UpdateActiveServers(len(spawners))
	utils.WriteJSON(w, http.StatusOK, spawners)
}

// GetSpawner returns a single spawner by numeric id.
func GetSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.GlobalRegistry.Get(id)
	if !ok {
		utils.WriteError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, s)
}

// DeleteSpawner removes a spawner from the registry.
func DeleteSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.GlobalRegistry.Delete(id) {
		utils.WriteError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	registry.GlobalStats.UpdateActiveServers(len(registry.GlobalRegistry.List()))
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// SpawnInstance triggers a new game instance on the specified spawner.
func SpawnInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	id, err := strconv.Atoi(rawID)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid spawner ID")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "spawn", nil, 30*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		var resData struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(resp.Data, &resData) == nil && resData.ID != "" {
			database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
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
	_, _ = w.Write(resp.Data)
}

// GetSpawnerLogs fetches and returns the log file content from a spawner.
func GetSpawnerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "get_logs", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ClearSpawnerLogs truncates the log file on a spawner.
func ClearSpawnerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "clear_logs", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ListSpawnerInstances fetches active game instances from a spawner.
func ListSpawnerInstances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "list_instances", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// UpdateSpawnerTemplate triggers the spawner to re-download the latest game server files from master.
func UpdateSpawnerTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "update_template", nil, 300*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// UpdateSpawnerInstance triggers an update (reinstall files) for a specific game instance.
func UpdateSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "update_instance", map[string]string{"instance_id": instanceID}, 300*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
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
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		NewID string `json:"new_id"`
	}
	if err := utils.DecodeJSON(r, &reqBody); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "rename_instance", map[string]string{"instance_id": instanceID, "new_id": reqBody.NewID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
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
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "remove_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
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
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "stop_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "stop",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "instance stopped", "id": instanceID})
}

// StartSpawnerInstance starts a specific game instance on a spawner.
func StartSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]

	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "start_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			SpawnerID:  id,
			InstanceID: instanceID,
			Action:     "start",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "instance started", "id": instanceID})
}

// RestartSpawnerInstance restarts a specific game instance on a spawner.
func RestartSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "restart_instance", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
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
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "get_instance_logs", map[string]string{"instance_id": instanceID}, 30*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
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
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "clear_instance_logs", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceStats proxies the stats request for a specific game instance.
func GetInstanceStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "get_instance_stats", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceHistory proxies the history request for a specific game instance.
func GetInstanceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "get_instance_history", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// GetInstanceHistoryActions retrieves the recorded action history for an instance.
func GetInstanceHistoryActions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	if database.DBConn == nil {
		utils.WriteJSON(w, http.StatusOK, []models.InstanceAction{})
		return
	}

	actions, err := database.GetInstanceActions(database.DBConn, id, instanceID)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to retrieve history")
		return
	}
	utils.WriteJSON(w, http.StatusOK, actions)
}

// BackupSpawnerInstance creates a backup of a game instance on a spawner.
func BackupSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "backup_instance", map[string]string{"instance_id": instanceID}, 300*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// RestoreSpawnerInstance restores a backup of a game instance on a spawner.
func RestoreSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		Filename string `json:"filename"`
	}
	if err := utils.DecodeJSON(r, &reqBody); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "restore_instance", map[string]string{"instance_id": instanceID, "filename": reqBody.Filename}, 300*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// ListSpawnerBackups lists backups of a game instance on a spawner.
func ListSpawnerBackups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "list_backups", map[string]string{"instance_id": instanceID}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// DeleteSpawnerBackup deletes a backup of a game instance on a spawner.
func DeleteSpawnerBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	var reqBody struct {
		Filename string `json:"filename"`
	}
	if err := utils.DecodeJSON(r, &reqBody); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "delete_backup", map[string]string{"instance_id": instanceID, "filename": reqBody.Filename}, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Data)
}

// Health is a lightweight liveness endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// StatsAPI returns JSON statistics about the server for the dashboard.
func StatsAPI(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, registry.GlobalStats.GetStatsMap())
}

// ErrorsAPI returns the recent error log.
func ErrorsAPI(w http.ResponseWriter, r *http.Request) {
	errors := registry.GlobalStats.GetErrors()
	utils.WriteJSON(w, http.StatusOK, errors)
}

// ClearErrorsAPI clears the error log.
func ClearErrorsAPI(w http.ResponseWriter, r *http.Request) {
	registry.GlobalStats.ClearErrors()
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Error log cleared"})
}

// ListTablesHandler returns a list of database tables.
func ListTablesHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	tables, err := database.ListTables(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to list tables: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, tables)
}

// GetTableCountsHandler returns table row counts.
func GetTableCountsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	counts, err := database.GetTableCounts(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to get table counts: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, counts)
}

// BackupDatabaseHandler triggers a pg_dump and streams it to the client.
func BackupDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "DB_DSN not configured")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"backup-%d.sql\"", time.Now().Unix()))

	cmd := exec.Command("pg_dump", dbDSN)
	cmd.Stdout = w

	if err := cmd.Run(); err != nil {
		log.Printf("pg_dump error: %v", err)
	}
}

// RestartServerHandler attempts to restart the server.
func RestartServerHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Restarting server..."})

	// Execute in a goroutine to allow the response to flush
	go func() {
		time.Sleep(1 * time.Second)
		log.Println("Restart triggered via API...")
		os.Exit(0) // Process manager should restart it
	}()
}
