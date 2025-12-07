package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEHub manages Server-Sent Events connections.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan string]bool
}

// NewSSEHub creates a new SSE hub.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan string]bool),
	}
}

// HandleSSE handles incoming SSE connections.
func (h *SSEHub) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Register client
	h.mu.Lock()
	h.clients[clientChan] = true
	h.mu.Unlock()

	log.Println("SSE client connected")

	// Ensure channel is closed on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		h.mu.Unlock()
		close(clientChan)
		log.Println("SSE client disconnected")
	}()

	// Send initial data
	h.sendUpdate(clientChan, "stats")
	h.sendUpdate(clientChan, "spawners")

	// Loop to send data to the client
	for {
		select {
		case msg := <-clientChan:
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *SSEHub) Broadcast(msgType string, payload interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	})
	if err != nil {
		log.Printf("SSE Broadcast error: %v", err)
		return
	}
	msg := string(data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client <- msg:
		default:
			// Skip if channel is full (slow client)
		}
	}
}

// Run starts the background ticker for updates.
func (h *SSEHub) Run() {
	statsTicker := time.NewTicker(1 * time.Second)
	spawnersTicker := time.NewTicker(2 * time.Second)
	defer statsTicker.Stop()
	defer spawnersTicker.Stop()

	for {
		select {
		case <-statsTicker.C:
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
			h.Broadcast("stats", stats)

		case <-spawnersTicker.C:
			spawners := registry.List()
			h.Broadcast("spawners", spawners)
		}
	}
}

// Helper to send a specific update type to a single client
func (h *SSEHub) sendUpdate(client chan string, msgType string) {
	var payload interface{}
	if msgType == "stats" {
		totalReq, totalErr, active, dbOK, uptime, mem, tx, rx := GlobalStats.GetStats()
		payload = map[string]interface{}{
			"uptime":          uptime.Milliseconds(),
			"active_spawners": active,
			"total_requests":  totalReq,
			"total_errors":    totalErr,
			"db_connected":    dbOK,
			"memory_usage":    mem,
			"bytes_sent":      tx,
			"bytes_received":  rx,
		}
	} else if msgType == "spawners" {
		payload = registry.List()
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	})
	
	select {
	case client <- string(data):
	default:
	}
}
