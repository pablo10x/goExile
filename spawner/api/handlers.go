package api

import (
	"log/slog"
	"net/http"
	"spawner/internal/config"
	"spawner/internal/game"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager *game.Manager
	config  *config.Config
	logger  *slog.Logger
}

func NewHandler(m *game.Manager, c *config.Config, l *slog.Logger) *Handler {
	return &Handler{manager: m, config: c, logger: l}
}

// RegisterRoutes sets up the API endpoints
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/spawn", h.HandleSpawn)
	router.GET("/instances", h.HandleListInstances)
	router.DELETE("/instance/:id", h.HandleStopInstance)
	router.GET("/health", h.HandleHealth)
}

func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "online",
		"region": h.config.Region,
		"uptime": "OK", // Could add real uptime
	})
}

func (h *Handler) HandleSpawn(c *gin.Context) {
	instance, err := h.manager.Spawn(c.Request.Context())
	if err != nil {
		h.logger.Error("Spawn failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

func (h *Handler) HandleListInstances(c *gin.Context) {
	instances := h.manager.ListInstances()
	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

func (h *Handler) HandleStopInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.manager.StopInstance(id); err != nil {
		h.logger.Error("Failed to stop instance", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance stopped", "id": id})
}