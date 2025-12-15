package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Validate API Key instead
	},
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
	mu              sync.RWMutex
	connections     map[int]*SpawnerConnection
	Broadcast       chan []byte
	Register        chan *SpawnerConnection
	Unregister      chan *SpawnerConnection
	pendingRequests map[string]chan WSResponse
	pendingMu       sync.Mutex
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
		connections:     make(map[int]*SpawnerConnection),
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
			manager.mu.Lock()
			// If exists, close old one
			if old, ok := manager.connections[conn.ID]; ok {
				close(old.WriteChan)
				old.Conn.Close()
			}
			manager.connections[conn.ID] = conn
			manager.mu.Unlock()
			registry.UpdateSpawnerStatus(conn.ID, "Online")
			log.Printf("🔌 Spawner #%d connected via WebSocket", conn.ID)

		case conn := <-manager.Unregister:
			manager.mu.Lock()
			if _, ok := manager.connections[conn.ID]; ok {
				delete(manager.connections, conn.ID)
				close(conn.WriteChan)
			}
			manager.mu.Unlock()
			registry.UpdateSpawnerStatus(conn.ID, "Offline")
			log.Printf("🔌 Spawner #%d disconnected", conn.ID)

		case message := <-manager.Broadcast:
			manager.mu.RLock()
			for _, conn := range manager.connections {
				select {
				case conn.WriteChan <- message:
				default:
					close(conn.WriteChan)
					delete(manager.connections, conn.ID)
				}
			}
			manager.mu.RUnlock()
		}
	}
}

// SendCommandSync sends a command and waits for a response.
func (manager *WSManager) SendCommandSync(spawnerID int, msgType string, payload interface{}, timeout time.Duration) (WSResponse, error) {
	// Check status first
	if s, ok := registry.Get(spawnerID); !ok {
		return WSResponse{}, fmt.Errorf("spawner not found")
	} else if s.Status != "Online" {
		return WSResponse{Status: "error", Error: "spawner not online"}, nil
	}

	manager.mu.RLock()
	conn, ok := manager.connections[spawnerID]
	manager.mu.RUnlock()

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
	if s, ok := registry.Get(spawnerID); !ok {
		return fmt.Errorf("spawner not found")
	} else if s.Status != "Online" {
		return fmt.Errorf("spawner not online")
	}

	manager.mu.RLock()
	conn, ok := manager.connections[spawnerID]
	manager.mu.RUnlock()

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
		var s Spawner
		if err := json.Unmarshal(msg.Payload, &s); err == nil {
			// Register spawner with full metadata and get assigned ID
			id, err := registry.Register(&s)
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
			log.Printf("✅ Spawner #%d Registered via WS (region=%s, host=%s, port=%d)", c.ID, s.Region, s.Host, s.Port)

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
			registry.UpdateHeartbeat(c.ID, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion)
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
