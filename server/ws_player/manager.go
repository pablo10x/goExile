package ws_player

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"exile/server/database"

	"github.com/gorilla/websocket"
)

// WSMessage represents the standard envelope for player websocket communication.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

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
		_ = existing.Conn.Close()
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
		_ = c.Conn.Close()
		log.Printf("PlayerWS: Player %d disconnected", c.PlayerID)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("PlayerWS: Invalid message format from %d: %v", c.PlayerID, err)
			continue
		}

		pm.handleMessage(c, wsMsg)
	}
}

func (pm *PlayerWSManager) handleMessage(c *PlayerConnection, msg WSMessage) {
	switch msg.Type {
	case "FRIEND_REQUEST_SEND":
		var payload struct {
			ReceiverID int64 `json:"receiver_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		pm.handleFriendRequest(c, payload.ReceiverID)

	case "FRIEND_REQUEST_ACCEPT":
		var payload struct {
			SenderID int64 `json:"sender_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		pm.handleFriendAccept(c, payload.SenderID)

	case "FRIEND_REQUEST_REJECT":
		var payload struct {
			SenderID int64 `json:"sender_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		// Logic: just delete the request
		_ = database.DeleteFriendRequest(database.DBConn, payload.SenderID, c.PlayerID)

	case "FRIEND_REMOVE":
		var payload struct {
			FriendID int64 `json:"friend_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		_ = database.RemoveFriendship(database.DBConn, c.PlayerID, payload.FriendID)

	default:
		log.Printf("PlayerWS: Unknown message type %s from %d", msg.Type, c.PlayerID)
	}
}

func (pm *PlayerWSManager) handleFriendRequest(c *PlayerConnection, receiverID int64) {
	if err := database.SendFriendRequest(database.DBConn, c.PlayerID, receiverID); err != nil {
		pm.sendError(c.PlayerID, fmt.Sprintf("Friend request failed: %v", err))
		return
	}

	// Notify receiver if online
	sender, _ := database.GetPlayerByID(database.DBConn, c.PlayerID)
	pm.SendMessage(receiverID, WSMessage{
		Type: "NOTIFY_FRIEND_REQUEST",
		Payload: mustMarshal(map[string]interface{}{
			"sender_id":   c.PlayerID,
			"sender_name": sender.Name,
		}),
	})
}

func (pm *PlayerWSManager) handleFriendAccept(c *PlayerConnection, senderID int64) {
	if err := database.AcceptFriendRequest(database.DBConn, senderID, c.PlayerID); err != nil {
		pm.sendError(c.PlayerID, fmt.Sprintf("Failed to accept friend: %v", err))
		return
	}

	// Notify both parties
	me, _ := database.GetPlayerByID(database.DBConn, c.PlayerID)
	them, _ := database.GetPlayerByID(database.DBConn, senderID)

	pm.SendMessage(c.PlayerID, WSMessage{
		Type: "NOTIFY_FRIEND_ACCEPTED",
		Payload: mustMarshal(map[string]interface{}{
			"friend_id":   senderID,
			"friend_name": them.Name,
		}),
	})

	pm.SendMessage(senderID, WSMessage{
		Type: "NOTIFY_FRIEND_ACCEPTED",
		Payload: mustMarshal(map[string]interface{}{
			"friend_id":   c.PlayerID,
			"friend_name": me.Name,
		}),
	})
}

func (pm *PlayerWSManager) sendError(playerID int64, message string) {
	pm.SendMessage(playerID, WSMessage{
		Type: "ERROR",
		Payload: mustMarshal(map[string]string{
			"message": message,
		}),
	})
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func (c *PlayerConnection) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.WriteChan:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
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

// IsPlayerOnline checks if a player is currently connected
func (pm *PlayerWSManager) IsPlayerOnline(playerID int64) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.Connections[playerID]
	return exists
}
