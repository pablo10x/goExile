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

// RegisterNode accepts registration from a models.Node service.
func RegisterNode(w http.ResponseWriter, r *http.Request) {
	var s models.Node
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

	registry.GlobalStats.UpdateActiveNodes(len(registry.GlobalRegistry.List()))
	utils.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatNode refreshes the LastSeen timestamp and stats for the given node ID.
func HeartbeatNode(w http.ResponseWriter, r *http.Request) {
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
		IsDraining       bool    `json:"is_draining"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := registry.GlobalRegistry.UpdateHeartbeat(id, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion, req.IsDraining); err != nil {
		utils.WriteError(w, r, http.StatusNotFound, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListNodes returns a JSON array with all currently registered nodes.
func ListNodes(w http.ResponseWriter, r *http.Request) {
	nodes := registry.GlobalRegistry.List()
	registry.GlobalStats.UpdateActiveNodes(len(nodes))
	utils.WriteJSON(w, http.StatusOK, nodes)
}

// GetNode returns a single node by numeric id.
func GetNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.GlobalRegistry.Get(id)
	if !ok {
		utils.WriteError(w, r, http.StatusNotFound, "node not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, s)
}

// DeleteNode removes a node from the registry.
func DeleteNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.GlobalRegistry.Delete(id) {
		utils.WriteError(w, r, http.StatusNotFound, "node not found")
		return
	}
	registry.GlobalStats.UpdateActiveNodes(len(registry.GlobalRegistry.List()))
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// UpdateNodeSettings updates the configuration for a node.
func UpdateNodeSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Region            string `json:"region"`
		MaxInstances      int    `json:"max_instances"`
		IsDraining        bool   `json:"is_draining"`
		Tags              string `json:"tags"`
		MaintenanceWindow string `json:"maintenance_window"`
		ResourceLimits    string `json:"resource_limits"`
		PublicIP          string `json:"public_ip"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.MaxInstances < 0 {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid max_instances")
		return
	}

	// Update in Registry/DB
	s, ok := registry.GlobalRegistry.Get(id)
	if !ok {
		utils.WriteError(w, r, http.StatusNotFound, "node not found")
		return
	}

	s.Region = req.Region
	s.MaxInstances = req.MaxInstances
	s.IsDraining = req.IsDraining
	s.Tags = req.Tags
	s.MaintenanceWindow = req.MaintenanceWindow
	s.ResourceLimits = req.ResourceLimits
	s.PublicIP = req.PublicIP

	// Persist to DB
	if database.DBConn != nil {
		if _, err := database.SaveNode(database.DBConn, s); err != nil {
			utils.WriteError(w, r, http.StatusInternalServerError, "failed to save node settings")
			return
		}
	}

	// Send update to Node via WS
	// The Node agent needs to support these new fields in the update_config payload
	err = ws.GlobalWSManager.SendCommand(id, "update_config", map[string]interface{}{
		"region":             req.Region,
		"max_instances":      req.MaxInstances,
		"is_draining":        req.IsDraining,
		"tags":               req.Tags,
		"maintenance_window": req.MaintenanceWindow,
		"resource_limits":    req.ResourceLimits,
		"public_ip":          req.PublicIP,
	})

	response := map[string]string{"message": "settings updated"}
	if err != nil {
		// Log warning and return it to the client so they know the node might not be updated yet
		log.Printf("Warning: Failed to push config update to node %d: %v", id, err)
		response["warning"] = "Database updated, but failed to push real-time update to node. Changes will apply on next restart/heartbeat."
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// SpawnNodeInstance triggers a new game instance on the specified node.
func SpawnNodeInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	id, err := strconv.Atoi(rawID)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid node ID")
		return
	}

	// Check if node is draining
	s, ok := registry.GlobalRegistry.Get(id)
	if !ok {
		utils.WriteError(w, r, http.StatusNotFound, "node not found")
		return
	}

	if s.IsDraining {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Node is in Drain Mode (Maintenance). No new instances allowed.")
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "spawn", nil, 30*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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
				NodeID:  id,
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

// GetNodeLogs fetches and returns the log file content from a node.
func GetNodeLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "get_logs", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// ClearNodeLogs truncates the log file on a node.
func ClearNodeLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "clear_logs", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// ListNodeInstances fetches active game instances from a node.
func ListNodeInstances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "list_instances", nil, 10*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// UpdateNodeTemplate triggers the node to re-download the latest game server files from master.
func UpdateNodeTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := ws.GlobalWSManager.SendCommandSync(id, "update_template", nil, 300*time.Second)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// UpdateNodeInstance triggers an update (reinstall files) for a specific game instance.
func UpdateNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
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

// RenameNodeInstance renames a specific game instance on a node.
func RenameNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
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

// RemoveNodeInstance removes a specific game instance on a node.
func RemoveNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
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

// StopNodeInstance stops a specific game instance on a node.
func StopNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
			InstanceID: instanceID,
			Action:     "stop",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "instance stopped", "id": instanceID})
}

// StartNodeInstance starts a specific game instance on a node.
func StartNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
			InstanceID: instanceID,
			Action:     "start",
			Timestamp:  time.Now().UTC(),
			Status:     "success",
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "instance started", "id": instanceID})
}

// RestartNodeInstance restarts a specific game instance on a node.
func RestartNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
		return
	}

	if resp.Status == "error" {
		utils.WriteError(w, r, http.StatusInternalServerError, resp.Error)
		return
	}

	if database.DBConn != nil {
		database.SaveInstanceAction(database.DBConn, &models.InstanceAction{
			NodeID:  id,
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// BackupNodeInstance creates a backup of a game instance on a node.
func BackupNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// RestoreNodeInstance restores a backup of a game instance on a node.
func RestoreNodeInstance(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// ListNodeBackups lists backups of a game instance on a node.
func ListNodeBackups(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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

// DeleteNodeBackup deletes a backup of a game instance on a node.
func DeleteNodeBackup(w http.ResponseWriter, r *http.Request) {
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
		utils.WriteError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact node via WS: %v", err))
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
