package ws_player

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// PlayerConnection represents a connected player client.
type PlayerConnection struct {
	Conn      *websocket.Conn
	PlayerID  int64
	UID       string
	WriteChan chan []byte
}

// PlayerWSManager manages all player websocket connections.
type PlayerWSManager struct {
	mu          sync.RWMutex
	Connections map[int64]*PlayerConnection // PlayerID -> Connection
	SessionKeys map[string]int64            // SessionKey -> PlayerID (temporary auth)
}

var GlobalPlayerWS *PlayerWSManager

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for game clients
	},
}

func InitPlayerWS() {
	GlobalPlayerWS = &PlayerWSManager{
		Connections: make(map[int64]*PlayerConnection),
		SessionKeys: make(map[string]int64),
	}
}

// RegisterSession creates a temporary session key for a player to connect.
func (pm *PlayerWSManager) RegisterSession(playerID int64, key string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.SessionKeys[key] = playerID
	
	// Auto-expire session key after 1 minute (client should connect immediately)
	go func(k string) {
		time.Sleep(1 * time.Minute)
		pm.mu.Lock()
		delete(pm.SessionKeys, k)
		pm.mu.Unlock()
	}(key)
}

func (pm *PlayerWSManager) HandleWS(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Session Key from Query Param
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing auth key", http.StatusUnauthorized)
		return
	}

	pm.mu.RLock()
	playerID, ok := pm.SessionKeys[key]
	pm.mu.RUnlock()

	if !ok {
		http.Error(w, "invalid or expired key", http.StatusUnauthorized)
		return
	}

	// 2. Upgrade Connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("PlayerWS: Upgrade error: %v", err)
		return
	}

	// 3. Register Connection
	client := &PlayerConnection{
		Conn:      conn,
		PlayerID:  playerID,
		WriteChan: make(chan []byte, 256),
	}

	pm.mu.Lock()
	// Close existing connection if any (kick old session)
	if existing, exists := pm.Connections[playerID]; exists {
		existing.Conn.Close()
	}
	pm.Connections[playerID] = client
	// Invalidate session key (one-time use)
	delete(pm.SessionKeys, key)
	pm.mu.Unlock()

	log.Printf("PlayerWS: Player %d connected", playerID)

	// Start pumps
	go client.writePump()
	go client.readPump(pm)
}

func (c *PlayerConnection) readPump(pm *PlayerWSManager) {
	defer func() {
		pm.mu.Lock()
		if current, ok := pm.Connections[c.PlayerID]; ok && current == c {
			delete(pm.Connections, c.PlayerID)
		}
		pm.mu.Unlock()
		c.Conn.Close()
		log.Printf("PlayerWS: Player %d disconnected", c.PlayerID)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Handle incoming messages (e.g. chat, party invites)
		// For now, we just log them.
		log.Printf("PlayerWS: Recv from %d: %s", c.PlayerID, string(message))
	}
}

func (c *PlayerConnection) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.WriteChan:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a JSON payload to a specific player
func (pm *PlayerWSManager) SendMessage(playerID int64, payload interface{}) {
	pm.mu.RLock()
	client, ok := pm.Connections[playerID]
	pm.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("PlayerWS: JSON marshal error: %v", err)
		return
	}

	select {
	case client.WriteChan <- data:
	default:
		// Buffer full, drop message or handle overflow
	}
}
