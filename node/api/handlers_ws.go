package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Only allow local connections for game server stats
		host := r.Host
		if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "127.0.0.1:") || strings.HasPrefix(host, "localhost:") {
			return true
		}
		return false
	},
}

// HandleInstanceWebSocket handles the WebSocket connection for a specific instance.
func (h *Handler) HandleInstanceWebSocket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket", "id", id, "error", err)
		return
	}
	defer func() { _ = conn.Close() }()

	h.logger.Info("Game server connected via WebSocket", "id", id)

	for {
		var msg struct {
			Type        string `json:"type"`
			PlayerCount int    `json:"player_count"`
			MaxPlayers  int    `json:"max_players"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error", "id", id, "error", err)
			}
			break
		}

		if msg.Type == "stats" {
			if err := h.manager.UpdatePlayerStats(id, msg.PlayerCount, msg.MaxPlayers); err != nil {
				h.logger.Warn("Failed to update player stats", "id", id, "error", err)
			}
		}
	}

	h.logger.Info("Game server WebSocket disconnected", "id", id)
}
