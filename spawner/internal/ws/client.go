package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"spawner/internal/config"
	"spawner/internal/game"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Client struct {
	config  *config.Config
	manager *game.Manager
	logger  *slog.Logger
	conn    *websocket.Conn
	send    chan []byte
	id      int
}

type WSMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

func NewClient(cfg *config.Config, m *game.Manager, l *slog.Logger) *Client {
	return &Client{
		config:  cfg,
		manager: m,
		logger:  l,
		send:    make(chan []byte, 256),
		id:      0, // Will be set after registration
	}
}

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
			c.conn.Close()
			continue
		}

		// Start loops
		go c.writePump()
		go c.heartbeatLoop()

		// Blocking read
		c.readPump()

		c.logger.Warn("WebSocket disconnected, reconnecting...")
		time.Sleep(3 * time.Second)
	}
}

func (c *Client) connect() error {
	headers := http.Header{}
	if c.config.MasterAPIKey != "" {
		headers.Set("X-API-Key", c.config.MasterAPIKey)
	}

	// Normalize URL: http://localhost:8081 -> ws://localhost:8081/api/spawners/ws
	// Assume MasterURL is http base
	url := c.config.MasterURL + "/api/spawners/ws"
	if len(url) > 4 && url[:4] == "http" {
		url = "ws" + url[4:]
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) sendRegister() error {
	port, _ := strconv.Atoi(c.config.Port)
	maxInstances := c.config.MaxGamePort - c.config.MinGamePort
	if maxInstances < 1 {
		maxInstances = 1
	}

	payload := map[string]interface{}{
		"region":            c.config.Region,
		"host":              c.config.Host,
		"port":              port,
		"max_instances":     maxInstances,
		"current_instances": len(c.manager.ListInstances()),
		"status":            "active",
	}
	data, _ := json.Marshal(payload)
	msg := WSMessage{Type: "REGISTER", Payload: data}
	bytes, _ := json.Marshal(msg)
	return c.conn.WriteMessage(websocket.TextMessage, bytes)
}

func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("Read error", "error", err)
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	// Simple channel consumer, relies on c.conn being open
	// In a robust impl, this would handle closes better
	// For now, if readPump exits, we reconnect, so this goroutine will leak or die?
	// We need a quit channel.
	// Simplified: We assume readPump controls the connection lifecycle.
}

func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		maxInstances := c.config.MaxGamePort - c.config.MinGamePort

		// Collect System Metrics
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

		// Read dynamic game version
		currentGameVersion := ""
		versionFile := filepath.Join(c.config.GameInstallDir, "version.txt")
		if content, err := os.ReadFile(versionFile); err == nil {
			currentGameVersion = string(bytes.TrimSpace(content))
		}

		status := "active"
		if c.manager.IsBusy() {
			status = "Updating"
		}

		payload := map[string]interface{}{
			"current_instances": len(c.manager.ListInstances()),
			"max_instances":     maxInstances,
			"status":            status,
			"cpu_usage":         cpuUsage,
			"mem_used":          memUsed,
			"mem_total":         memTotal,
			"disk_used":         diskUsed,
			"disk_total":        diskTotal,
			"game_version":      currentGameVersion,
		}
		data, _ := json.Marshal(payload)
		msg := WSMessage{Type: "HEARTBEAT", Payload: data}
		bytes, _ := json.Marshal(msg)

		if err := c.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
			return
		}
	}
}

func (c *Client) handleMessage(msg WSMessage) {
	c.logger.Info("Received message", "type", msg.Type, "req_id", msg.RequestID)

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
				c.logger.Info("✅ Successfully registered with Master Server via WebSocket", "spawner_id", c.id)
			} else {
				c.logger.Error("❌ Registration failed", "error", resp.Error)
				// Close connection to trigger reconnection
				c.conn.Close()
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
		json.Unmarshal(msg.Payload, &req)
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
		json.Unmarshal(msg.Payload, &req)
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
		json.Unmarshal(msg.Payload, &req)

		c.manager.StopInstance(req.InstanceID)
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
		json.Unmarshal(msg.Payload, &req)
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
	msg := WSMessage{Type: "RESPONSE", Payload: payload}
	bytes, _ := json.Marshal(msg)
	c.conn.WriteMessage(websocket.TextMessage, bytes)
}
