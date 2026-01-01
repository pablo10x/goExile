// Package ws provides the WebSocket client for communicating with the master server.
package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"node/internal/config"
	"node/internal/game"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Client manages the WebSocket connection to the Master Server.
type Client struct {
	config    *config.Config
	manager   *game.Manager
	logger    *slog.Logger
	conn      *websocket.Conn
	send      chan []byte
	id        int
	metrics   cachedMetrics
	metricsMu sync.RWMutex
}

type cachedMetrics struct {
	CPUUsage  float64
	MemUsed   uint64
	MemTotal  uint64
	DiskUsed  uint64
	DiskTotal uint64
}

// Message represents a message sent over the WebSocket connection.
type Message struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// NewClient creates a new WebSocket client.
func NewClient(cfg *config.Config, m *game.Manager, l *slog.Logger) *Client {
	return &Client{
		config:  cfg,
		manager: m,
		logger:  l,
		send:    make(chan []byte, 256),
		id:      0, // Will be set after registration
	}
}

// Start initiates the WebSocket client loop.
func (c *Client) Start() {
	for {
		if err := c.connect(); err != nil {
			c.logger.Error("WebSocket connection failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		c.logger.Info("✅ Connected to Master Server via WebSocket")

		// Send Register
		if err := c.sendRegister(); err != nil {
			c.logger.Error("Failed to send register", "error", err)
			_ = c.conn.Close()
			continue
		}

		// Create a done channel to signal shutdown of goroutines for this connection
		done := make(chan struct{})

		// Start loops
		go c.writePump(done)
		go c.heartbeatLoop(done)
		go c.metricsLoop(done) // Start async metrics collection

		// Blocking read
		c.readPump()

		// Signal other goroutines to stop
		close(done)

		c.logger.Warn("WebSocket disconnected, reconnecting...")
		time.Sleep(3 * time.Second)
	}
}

func (c *Client) connect() error {
	headers := http.Header{}
	if c.config.MasterAPIKey != "" {
		headers.Set("X-API-Key", c.config.MasterAPIKey)
	}

	// Normalize URL: http://localhost:8081 -> ws://localhost:8081/api/nodes/ws
	// Assume MasterURL is http base
	baseURL := strings.TrimSuffix(c.config.MasterURL, "/")
	url := baseURL + "/api/nodes/ws"
	if len(url) > 4 && url[:4] == "http" {
		url = "ws" + url[4:]
	}

	conn, resp, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		if resp != nil {
			c.logger.Error("WebSocket connection failed", "error", err, "status_code", resp.StatusCode)
		} else {
			c.logger.Error("WebSocket connection failed", "error", err)
		}
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) sendRegister() error {
	port, _ := strconv.Atoi(c.config.Port)
	maxInstances := c.config.MaxInstances
	if maxInstances < 1 {
		maxInstances = 1
	}

	payload := map[string]interface{}{
		"region":            c.config.Region,
		"host":              c.config.Host,
		"port":              port,
		"max_instances":     maxInstances,
		"current_instances": len(c.manager.ListInstances()),
		"status":            "Online",
	}
	data, _ := json.Marshal(payload)
	msg := Message{Type: "REGISTER", Payload: data}
	bytes, _ := json.Marshal(msg)
	return c.conn.WriteMessage(websocket.TextMessage, bytes)
}

func (c *Client) readPump() {
	defer func() { _ = c.conn.Close() }()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("Read error", "error", err)
			return
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) writePump(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case message := <-c.send:
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Error("Write error", "error", err)
				_ = c.conn.Close() // Force readPump to exit
				return
			}
		}
	}
}

// metricsLoop collects system stats asynchronously to prevent blocking the heartbeat
func (c *Client) metricsLoop(done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial collection
	c.collectMetrics()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			c.collectMetrics()
			c.rotateLogs()
		}
	}
}

func (c *Client) rotateLogs() {
	const maxLogSize = 5 * 1024 * 1024 // 5MB
	const keepLines = 20

	info, err := os.Stat("node.log")
	if err != nil {
		if !os.IsNotExist(err) {
			c.logger.Error("Failed to stat log file", "error", err)
		}
		return
	}

	if info.Size() < maxLogSize {
		return
	}

	c.logger.Info("Rotating log file", "current_size", info.Size())

	content, err := os.ReadFile("node.log")
	if err != nil {
		c.logger.Error("Failed to read log file for rotation", "error", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > keepLines {
		newContent := strings.Join(lines[len(lines)-keepLines:], "\n")
		if err := os.WriteFile("node.log", []byte(newContent), 0600); err != nil {
			c.logger.Error("Failed to write rotated log file", "error", err)
		} else {
			c.logger.Info("Log file rotated", "new_lines", keepLines)
		}
	}
}

func (c *Client) collectMetrics() {
	var cpuUsage float64
	if percentages, err := cpu.Percent(0, false); err == nil && len(percentages) > 0 {
		cpuUsage = percentages[0]
	}

	var memUsed, memTotal uint64
	if v, err := mem.VirtualMemory(); err == nil {
		memUsed = v.Used
		memTotal = v.Total
	}

	var diskUsed, diskTotal uint64
	// Use current directory for disk usage
	if d, err := disk.Usage("."); err == nil {
		diskUsed = d.Used
		diskTotal = d.Total
	}

	c.metricsMu.Lock()
	c.metrics = cachedMetrics{
		CPUUsage:  cpuUsage,
		MemUsed:   memUsed,
		MemTotal:  memTotal,
		DiskUsed:  diskUsed,
		DiskTotal: diskTotal,
	}
	c.metricsMu.Unlock()
}

func (c *Client) heartbeatLoop(done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			maxInstances := c.config.MaxInstances

			// Use cached metrics
			c.metricsMu.RLock()
			metrics := c.metrics
			c.metricsMu.RUnlock()

			// Read dynamic game version
			currentGameVersion := ""
			versionFile := filepath.Join(c.config.GameInstallDir, "version.txt")
			if content, err := os.ReadFile(versionFile); err == nil {
				currentGameVersion = string(bytes.TrimSpace(content))
			}

			status := "Online"
			if c.manager.IsBusy() {
				status = "Updating"
			}

			payload := map[string]interface{}{
				"current_instances": len(c.manager.ListInstances()),
				"max_instances":     maxInstances,
				"status":            status,
				"cpu_usage":         metrics.CPUUsage,
				"mem_used":          metrics.MemUsed,
				"mem_total":         metrics.MemTotal,
				"disk_used":         metrics.DiskUsed,
				"disk_total":        metrics.DiskTotal,
				"game_version":      currentGameVersion,
				"is_draining":       c.config.IsDraining,
			}
			data, _ := json.Marshal(payload)
			msg := Message{Type: "HEARTBEAT", Payload: data}
			bytes, _ := json.Marshal(msg)

			c.logger.Debug("Sending heartbeat", "timestamp", time.Now().UnixMilli())

			select {
			case c.send <- bytes:
			case <-done:
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg Message) {
	if msg.Type != "get_instance_logs" && msg.Type != "get_logs" && msg.Type != "get_instance_stats" {
		c.logger.Info("Received message", "type", msg.Type, "req_id", msg.RequestID)
	}

	switch msg.Type {
	case "REGISTER_RESPONSE":
		var resp struct {
			Status string `json:"status"`
			ID     int    `json:"id"`
			Error  string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(msg.Payload, &resp); err == nil {
			if resp.Status == "success" {
				c.id = resp.ID
				c.logger.Info("✅ Successfully registered with Master Server via WebSocket", "node_id", c.id)
			} else {
				c.logger.Error("❌ Registration failed", "error", resp.Error)
				if strings.Contains(resp.Error, "not enrolled") {
					c.logger.Error("FATAL: Node is not enrolled. Please use the dashboard to generate an enrollment key and run with -key <key>.")
					os.Exit(1)
				}
				// Close connection to trigger reconnection
				_ = c.conn.Close()
			}
		}
	case "spawn":
		ctx := context.Background()
		inst, err := c.manager.Spawn(ctx)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(inst)
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "start_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.StartInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			c.sendResponse(msg.RequestID, "success", nil, "")
		}

	case "stop_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.StopInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			c.sendResponse(msg.RequestID, "success", nil, "")
		}

	case "restart_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}

		_ = c.manager.StopInstance(req.InstanceID) // Ignore stop error, it might already be stopped
		err := c.manager.StartInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			c.sendResponse(msg.RequestID, "success", nil, "")
		}

	case "remove_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.RemoveInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			c.sendResponse(msg.RequestID, "success", nil, "")
		}

	case "list_instances":
		instances := c.manager.ListInstances()
		data, _ := json.Marshal(map[string]interface{}{"instances": instances})
		c.sendResponse(msg.RequestID, "success", data, "")

	case "get_instance_stats":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.logger.Error("Failed to unmarshal get_instance_stats payload", "error", err)
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}

		stats, err := c.manager.GetInstanceStats(req.InstanceID)
		if err != nil {
			c.logger.Warn("Failed to get instance stats", "instance_id", req.InstanceID, "error", err)
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(stats)
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "get_instance_history":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		history, err := c.manager.GetInstanceHistory(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]interface{}{"history": history})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "update_config":
		var req struct {
			Region            string `json:"region"`
			MaxInstances      int    `json:"max_instances"`
			IsDraining        bool   `json:"is_draining"`
			Tags              string `json:"tags"`
			MaintenanceWindow string `json:"maintenance_window"`
			ResourceLimits    string `json:"resource_limits"`
			PublicIP          string `json:"public_ip"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.logger.Error("Failed to unmarshal update_config payload", "error", err)
			return
		}

		c.logger.Info("Received config update", "region", req.Region, "max_instances", req.MaxInstances, "is_draining", req.IsDraining)

		// Update runtime config
		if req.Region != "" {
			c.config.Region = req.Region
		}
		if req.MaxInstances > 0 {
			c.config.MaxInstances = req.MaxInstances
		}
		c.config.IsDraining = req.IsDraining
		c.config.Tags = req.Tags
		c.config.MaintenanceWindow = req.MaintenanceWindow
		c.config.ResourceLimits = req.ResourceLimits
		c.config.PublicIP = req.PublicIP

		// Persist to .env
		updates := map[string]string{}
		if req.Region != "" {
			updates["REGION"] = req.Region
		}
		if req.MaxInstances > 0 {
			updates["MAX_INSTANCES"] = strconv.Itoa(req.MaxInstances)
		}
		updates["IS_DRAINING"] = strconv.FormatBool(req.IsDraining)
		if req.Tags != "" {
			updates["TAGS"] = req.Tags
		}
		if req.MaintenanceWindow != "" {
			updates["MAINTENANCE_WINDOW"] = req.MaintenanceWindow
		}
		if req.ResourceLimits != "" {
			updates["RESOURCE_LIMITS"] = req.ResourceLimits
		}
		if req.PublicIP != "" {
			updates["PUBLIC_IP"] = req.PublicIP
		}

		if err := config.SaveConfigToEnv(updates); err != nil {
			c.logger.Error("Failed to save updated config to .env", "error", err)
		} else {
			c.logger.Info("Updated configuration saved to .env")
		}

	case "update_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.UpdateInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "instance updated"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "rename_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
			NewID      string `json:"new_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.RenameInstance(req.InstanceID, req.NewID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "instance renamed", "new_id": req.NewID})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "backup_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.BackupInstance(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "backup created"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "restore_instance":
		var req struct {
			InstanceID string `json:"instance_id"`
			Filename   string `json:"filename"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.RestoreInstance(req.InstanceID, req.Filename)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "instance restored"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "list_backups":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		backups, err := c.manager.ListBackups(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]interface{}{"backups": backups})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "delete_backup":
		var req struct {
			InstanceID string `json:"instance_id"`
			Filename   string `json:"filename"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.DeleteBackup(req.InstanceID, req.Filename)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "backup deleted"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "update_template":
		updatedVersion, err := c.manager.UpdateTemplate()
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			// Read local version
			versionFile := filepath.Join(c.config.GameInstallDir, "version.txt")
			localVersion := ""
			if content, err := os.ReadFile(versionFile); err == nil {
				localVersion = string(bytes.TrimSpace(content))
			}
			message := "Template updated."
			if localVersion == updatedVersion {
				message = "Template already up to date."
			}
			data, _ := json.Marshal(map[string]string{"message": message, "version": updatedVersion})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "get_logs":
		content, err := os.ReadFile("node.log")
		var size int64
		if info, sErr := os.Stat("node.log"); sErr == nil {
			size = info.Size()
		}

		if err != nil {
			if os.IsNotExist(err) {
				data, _ := json.Marshal(map[string]interface{}{"logs": "", "size": 0})
				c.sendResponse(msg.RequestID, "success", data, "")
			} else {
				c.sendResponse(msg.RequestID, "error", nil, "failed to read logs")
			}
		} else {
			data, _ := json.Marshal(map[string]interface{}{"logs": string(content), "size": size})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "clear_logs":
		if err := os.Truncate("node.log", 0); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "failed to clear logs")
		} else {
			data, _ := json.Marshal(map[string]string{"message": "logs cleared"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}

	case "get_instance_logs":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		logPath, err := c.manager.GetInstanceLogPath(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			content, err := os.ReadFile(logPath)
			if err != nil {
				c.sendResponse(msg.RequestID, "error", nil, "log file not found")
			} else {
				data, _ := json.Marshal(map[string]string{"logs": string(content)})
				c.sendResponse(msg.RequestID, "success", data, "")
			}
		}

	case "clear_instance_logs":
		var req struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			c.sendResponse(msg.RequestID, "error", nil, "invalid payload")
			return
		}
		err := c.manager.ClearInstanceLogs(req.InstanceID)
		if err != nil {
			c.sendResponse(msg.RequestID, "error", nil, err.Error())
		} else {
			data, _ := json.Marshal(map[string]string{"message": "logs cleared"})
			c.sendResponse(msg.RequestID, "success", data, "")
		}
	}
}

func (c *Client) sendResponse(reqID, status string, data json.RawMessage, errStr string) {
	resp := struct {
		RequestID string          `json:"request_id"`
		Status    string          `json:"status"`
		Data      json.RawMessage `json:"data,omitempty"`
		Error     string          `json:"error,omitempty"`
	}{
		RequestID: reqID,
		Status:    status,
		Data:      data,
		Error:     errStr,
	}

	payload, _ := json.Marshal(resp)
	msg := Message{Type: "RESPONSE", Payload: payload}
	bytes, _ := json.Marshal(msg)

	select {
	case c.send <- bytes:
	default:
		c.logger.Warn("Send buffer full, dropping response", "req_id", reqID)
	}
}
