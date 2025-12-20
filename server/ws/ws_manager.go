package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/registry"
	"exile/server/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Validate Origin header to prevent Cross-Site WebSocket Hijacking
		origin := r.Header.Get("Origin")

		// If no Origin header, this is likely a direct API call (not from browser)
		// Still require API key validation in the handler
		if origin == "" {
			return true
		}

		// Get allowed origins from environment or use defaults
		allowedOrigins := getAllowedOrigins()

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		log.Printf("WebSocket connection rejected: invalid origin %s", origin)
		return false
	},
}

// getAllowedOrigins returns the list of allowed WebSocket origins
func getAllowedOrigins() []string {
	// Try loading from DB first
	if database.DBConn != nil {
		if c, err := database.GetConfigByKey(database.DBConn, "allowed_origins"); err == nil && c != nil {
			return splitAndTrim(c.Value, ",")
		}
	}

	// Default allowed origins for development and production
	defaults := []string{
		"http://localhost:5173", // SvelteKit dev server
		"http://localhost:8081", // Backend server
		"http://127.0.0.1:5173",
		"http://127.0.0.1:8081",
	}

	// Add custom origins from environment variable if set
	// Format: ALLOWED_ORIGINS=https://example.com,https://admin.example.com
	if customOrigins := utils.GetEnv("ALLOWED_ORIGINS", ""); customOrigins != "" {
		for _, origin := range splitAndTrim(customOrigins, ",") {
			if origin != "" {
				defaults = append(defaults, origin)
			}
		}
	}

	return defaults
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// SpawnerConnection represents a connected Spawner.
type SpawnerConnection struct {
	ID        int
	Conn      *websocket.Conn
	WriteChan chan []byte
	Manager   *WSManager
}

// WSManager manages Spawner connections.
type WSManager struct {
	Mu              sync.RWMutex
	Connections     map[int]*SpawnerConnection
	Broadcast       chan []byte
	Register        chan *SpawnerConnection
	Unregister      chan *SpawnerConnection
	pendingRequests map[string]chan WSResponse
	pendingMu       sync.Mutex
}

// IsClientConnected checks if a spawner with the given ID is currently connected via WebSocket.
func (manager *WSManager) IsClientConnected(spawnerID int) bool {
	manager.Mu.RLock()
	defer manager.Mu.RUnlock()
	_, ok := manager.Connections[spawnerID]
	return ok
}

type WSMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type WSResponse struct {
	RequestID string          `json:"request_id"`
	Status    string          `json:"status"` // "success", "error"
	Data      json.RawMessage `json:"data,omitempty"`
	Error     string          `json:"error,omitempty"`
}

var GlobalWSManager = NewWSManager()

func NewWSManager() *WSManager {
	return &WSManager{
		Connections:     make(map[int]*SpawnerConnection),
		Broadcast:       make(chan []byte),
		Register:        make(chan *SpawnerConnection),
		Unregister:      make(chan *SpawnerConnection),
		pendingRequests: make(map[string]chan WSResponse),
	}
}

func (manager *WSManager) Run() {
	for {
		select {
		case conn := <-manager.Register:
			manager.Mu.Lock()
			// If exists, close old one
			if old, ok := manager.Connections[conn.ID]; ok {
				close(old.WriteChan)
				old.Conn.Close()
			}
			manager.Connections[conn.ID] = conn
			manager.Mu.Unlock()
			registry.GlobalRegistry.UpdateSpawnerStatus(conn.ID, "Online")

		case conn := <-manager.Unregister:
			manager.Mu.Lock()
			if _, ok := manager.Connections[conn.ID]; ok {
				delete(manager.Connections, conn.ID)
				close(conn.WriteChan)
			}
			manager.Mu.Unlock()
			registry.GlobalRegistry.UpdateSpawnerStatus(conn.ID, "Offline")

		case message := <-manager.Broadcast:
			manager.Mu.RLock()
			for _, conn := range manager.Connections {
				select {
				case conn.WriteChan <- message:
				default:
					close(conn.WriteChan)
					delete(manager.Connections, conn.ID)
				}
			}
			manager.Mu.RUnlock()
		}
	}
}

// SendCommandSync sends a command and waits for a response.
func (manager *WSManager) SendCommandSync(spawnerID int, msgType string, payload interface{}, timeout time.Duration) (WSResponse, error) {
	// Check status first
	if s, ok := registry.GlobalRegistry.Get(spawnerID); !ok {
		return WSResponse{}, fmt.Errorf("spawner not found")
	} else if s.Status != "Online" {
		return WSResponse{Status: "error", Error: "spawner not online"}, nil
	}

	manager.Mu.RLock()
	conn, ok := manager.Connections[spawnerID]
	manager.Mu.RUnlock()

	if !ok {
		return WSResponse{}, log.Output(2, "Spawner not connected")
	}

	reqID := strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Intn(1000))
	respChan := make(chan WSResponse, 1)

	manager.pendingMu.Lock()
	manager.pendingRequests[reqID] = respChan
	manager.pendingMu.Unlock()

	defer func() {
		manager.pendingMu.Lock()
		delete(manager.pendingRequests, reqID)
		manager.pendingMu.Unlock()
	}()

	data, err := json.Marshal(payload)
	if err != nil {
		return WSResponse{}, err
	}

	msg := WSMessage{
		Type:      msgType,
		RequestID: reqID,
		Payload:   data,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return WSResponse{}, err
	}

	select {
	case conn.WriteChan <- bytes:
	default:
		return WSResponse{}, log.Output(2, "Write channel full")
	}

	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(timeout):
		return WSResponse{}, log.Output(2, "Timeout waiting for response")
	}
}

// SendCommand sends a command to a specific Spawner asynchronously.
func (manager *WSManager) SendCommand(spawnerID int, msgType string, payload interface{}) error {
	// Check status first
	if s, ok := registry.GlobalRegistry.Get(spawnerID); !ok {
		return fmt.Errorf("spawner not found")
	} else if s.Status != "Online" {
		return fmt.Errorf("spawner not online")
	}

	manager.Mu.RLock()
	conn, ok := manager.Connections[spawnerID]
	manager.Mu.RUnlock()

	if !ok {
		return log.Output(2, "Spawner not connected")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := WSMessage{
		Type:    msgType,
		Payload: data,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.WriteChan <- bytes
	return nil
}

// HandleWS handles WebSocket requests from Spawners.
func (manager *WSManager) HandleWS(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate (Already checked by UnifiedAuthMiddleware if configured correctly)

	// 2. Upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &SpawnerConnection{
		Conn:      conn,
		WriteChan: make(chan []byte, 256),
		Manager:   manager,
	}

	go client.writePump()
	go client.readPump()
}

func (c *SpawnerConnection) readPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Refresh deadline on any message (e.g. Heartbeat)
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *SpawnerConnection) writePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.WriteChan:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *SpawnerConnection) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "REGISTER":
		var s models.Spawner
		if err := json.Unmarshal(msg.Payload, &s); err == nil {
			// Enforce Enrollment: Check if spawner exists
			existing, found := registry.GlobalRegistry.Lookup(s.Host, s.Port)
			if !found {
				registry.GlobalStats.RecordSecurityEvent("Unauthorized Spawner Connection", fmt.Sprintf("Host: %s, Port: %d", s.Host, s.Port), c.Conn.RemoteAddr().String())
				errorResp := WSMessage{
					Type: "REGISTER_RESPONSE",
					Payload: func() json.RawMessage {
						data, _ := json.Marshal(map[string]interface{}{
							"status": "error",
							"error":  "Spawner not enrolled. Please use the dashboard to generate an enrollment key.",
						})
						return data
					}(),
				}
				if bytes, err := json.Marshal(errorResp); err == nil {
					select {
					case c.WriteChan <- bytes:
					default:
					}
				}
				// Close connection after sending error
				go func() {
					time.Sleep(1 * time.Second)
					c.Manager.Unregister <- c
				}()
				return
			}

			// Maintain Identity
			s.ID = existing.ID
			s.Name = existing.Name

			// Register spawner (Update)
			id, err := registry.GlobalRegistry.Register(&s)
			if err != nil {
				log.Printf("❌ Failed to register spawner via WS: %v", err)
				// Send error response back to spawner
				errorResp := WSMessage{
					Type: "REGISTER_RESPONSE",
					Payload: func() json.RawMessage {
						data, _ := json.Marshal(map[string]interface{}{
							"status": "error",
							"error":  err.Error(),
						})
						return data
					}(),
				}
				if bytes, err := json.Marshal(errorResp); err == nil {
					select {
					case c.WriteChan <- bytes:
					default:
					}
				}
				return
			}
			c.ID = id
			c.Manager.Register <- c

			// Send success response with assigned ID
			successResp := WSMessage{
				Type: "REGISTER_RESPONSE",
				Payload: func() json.RawMessage {
					data, _ := json.Marshal(map[string]interface{}{
						"status": "success",
						"id":     id,
					})
					return data
				}(),
			}
			if bytes, err := json.Marshal(successResp); err == nil {
				select {
				case c.WriteChan <- bytes:
				default:
				}
			}
		} else {
			log.Printf("❌ Invalid REGISTER payload: %v", err)
		}
	case "HEARTBEAT":
		// Handle heartbeat payload
		// We can update the registry directly here
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
		if err := json.Unmarshal(msg.Payload, &req); err == nil {
			registry.GlobalRegistry.UpdateHeartbeat(c.ID, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion)
		}
	case "RESPONSE":
		var resp WSResponse
		if err := json.Unmarshal(msg.Payload, &resp); err == nil {
			c.Manager.pendingMu.Lock()
			if ch, ok := c.Manager.pendingRequests[resp.RequestID]; ok {
				select {
				case ch <- resp:
				default:
				}
			}
			c.Manager.pendingMu.Unlock()
		}
	case "LOGS":
		// Handle streaming logs?
	}
}
