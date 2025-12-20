package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow local connections
	},
}

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
	defer conn.Close()

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
