package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"exile/server/registry"
)

// Event represents a structured SSE event
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SSEHub manages Server-Sent Events connections.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan string]bool
}

// GlobalHub is the shared singleton for system-wide event broadcasting.
var GlobalHub = NewSSEHub()

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
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Register client
	h.mu.Lock()
	h.clients[clientChan] = true
	h.mu.Unlock()

	// Ensure channel is closed on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		h.mu.Unlock()
		close(clientChan)
	}()

	// Send initial data
	h.sendUpdate(clientChan, "stats")
	h.sendUpdate(clientChan, "nodes")

	// Loop to send data to the client
	for {
		select {
		case msg := <-clientChan:
			// Ensure msg doesn't contain newlines that break SSE protocol
			cleanMsg := strings.ReplaceAll(msg, "\n", "")
			_, err := fmt.Fprintf(w, "data: %s\n\n", cleanMsg)
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
func (h *SSEHub) Broadcast(event Event) {
	data, err := json.Marshal(event)
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
	nodesTicker := time.NewTicker(2 * time.Second)
	defer statsTicker.Stop()
	defer nodesTicker.Stop()

	for {
		select {
		case <-statsTicker.C:
			h.Broadcast(Event{Type: "stats", Payload: registry.GlobalStats.GetStatsMap()})

		case <-nodesTicker.C:
			nodes := registry.GlobalRegistry.List()
			h.Broadcast(Event{Type: "nodes", Payload: nodes})
		}
	}
}

// Helper to send a specific update type to a single client
func (h *SSEHub) sendUpdate(client chan string, msgType string) {
	var payload interface{}
	if msgType == "stats" {
		payload = registry.GlobalStats.GetStatsMap()
	} else if msgType == "nodes" {
		payload = registry.GlobalRegistry.List()
	}

	event := Event{
		Type:    msgType,
		Payload: payload,
	}
	data, _ := json.Marshal(event)

	select {
	case client <- string(data):
	default:
	}
}
