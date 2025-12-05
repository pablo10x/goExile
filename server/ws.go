package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketHub manages WebSocket connections and broadcasts updates.
type WebSocketHub struct {
	mu          sync.RWMutex
	clients     map[*WebSocketClient]bool
	broadcast   chan interface{}
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
}

// WebSocketClient represents a connected WebSocket client.
type WebSocketClient struct {
	hub  *WebSocketHub
	conn *websocket.Conn
	send chan interface{}
}

// WSMessage represents a message sent over WebSocket.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewWebSocketHub creates a new WebSocket hub.
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		broadcast:   make(chan interface{}, 256),
		register:    make(chan *WebSocketClient),
		unregister:  make(chan *WebSocketClient),
		clients:     make(map[*WebSocketClient]bool),
	}
}

// Run starts the hub's event loop.
func (h *WebSocketHub) Run() {
	statsTicket := time.NewTicker(1 * time.Second)
	serversTicket := time.NewTicker(2 * time.Second)

	defer statsTicket.Stop()
	defer serversTicket.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Println("WebSocket client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Client's send channel is full, skip
				}
			}
			h.mu.RUnlock()

		case <-statsTicket.C:
			// Broadcast stats update
			totalReq, totalErr, active, dbOK, uptime := GlobalStats.GetStats()
			stats := map[string]interface{}{
				"uptime":         uptime.Milliseconds(),
				"active_servers": active,
				"total_requests": totalReq,
				"total_errors":   totalErr,
				"db_connected":   dbOK,
			}
			msg := WSMessage{
				Type:    "stats",
				Payload: stats,
			}
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
				}
			}
			h.mu.RUnlock()

		case <-serversTicket.C:
			// Broadcast servers update
			servers := registry.List()
			msg := WSMessage{
				Type:    "servers",
				Payload: servers,
			}
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// var upgrader allows WebSocket upgrades.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development; restrict in production
		return true
	},
}

// HandleWebSocket upgrades an HTTP connection to WebSocket.
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		hub:  h,
		conn: conn,
		send: make(chan interface{}, 256),
	}

	h.register <- client

	// Start goroutines to handle reading and writing
	go client.readPump()
	go client.writePump()
}

// readPump reads messages from the WebSocket connection.
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			encoder := json.NewEncoder(w)
			if err := encoder.Encode(msg); err != nil {
				w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
