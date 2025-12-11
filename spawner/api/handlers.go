package api

import (
	"bufio"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"spawner/internal/config"
	"spawner/internal/game"
	"strings"
	"time"

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
	router.POST("/update-template", h.HandleUpdateTemplate)
	router.GET("/instances", h.HandleListInstances)
	router.POST("/instance/:id/stop", h.HandleStopInstance)
	router.DELETE("/instance/:id", h.HandleRemoveInstance)
	router.POST("/instance/:id/start", h.HandleStartInstance)
	router.POST("/instance/:id/restart", h.HandleRestartInstance)
	router.POST("/instance/:id/update", h.HandleUpdateInstance)
	router.POST("/instance/:id/rename", h.HandleRenameInstance)
	router.GET("/instance/:id/stats", h.HandleInstanceStats)
	router.GET("/instance/:id/stats/history", h.HandleInstanceHistory)
	router.GET("/instance/:id/logs", h.HandleInstanceLogs)
	router.DELETE("/instance/:id/logs", h.HandleClearInstanceLogs)
	router.GET("/health", h.HandleHealth)
	router.GET("/logs", h.HandleGetLogs)
	router.DELETE("/logs", h.HandleClearLogs)
	
	// Backup endpoints
	router.POST("/instance/:id/backup", h.HandleBackupInstance)
	router.POST("/instance/:id/restore", h.HandleRestoreInstance)
	router.GET("/instance/:id/backups", h.HandleListBackups)
	router.POST("/instance/:id/backup/delete", h.HandleDeleteBackup)
}

func (h *Handler) HandleUpdateTemplate(c *gin.Context) {
	updatedVersion, err := h.manager.UpdateTemplate()
	if err != nil {
		h.logger.Error("Failed to update template", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	localVersion := ""
	versionFile := filepath.Join(h.config.GameInstallDir, "version.txt")
	if content, err := os.ReadFile(versionFile); err == nil {
		localVersion = strings.TrimSpace(string(content))
	}

	message := "Template updated."
	if localVersion == updatedVersion {
		message = "Template already up to date."
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "version": updatedVersion})
}

func (h *Handler) HandleUpdateInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.manager.UpdateInstance(id); err != nil {
		h.logger.Error("Failed to update instance", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance updated"})
}

func (h *Handler) HandleRenameInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req struct {
		NewID string `json:"new_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.NewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_id is required"})
		return
	}

	if err := h.manager.RenameInstance(id, req.NewID); err != nil {
		h.logger.Error("Failed to rename instance", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance renamed", "new_id": req.NewID})
}

func (h *Handler) HandleRemoveInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.manager.RemoveInstance(id); err != nil {
		h.logger.Error("Failed to remove instance", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance removed", "id": id})
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

func (h *Handler) HandleStartInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	 err := h.manager.StartInstance(id)
	if err != nil {
		h.logger.Error("Failed to start instance", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,gin.H{"message": "instance started", "id": id})
}

func (h *Handler) HandleRestartInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

    // Try stop (ignore error if not running)
    h.manager.StopInstance(id)
    
    // Start
    if err := h.manager.StartInstance(id); err != nil {
        h.logger.Error("Failed to start instance during restart", "id", id, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	c.JSON(http.StatusOK, gin.H{"message": "instance restarted", "id": id})
}

func (h *Handler) HandleInstanceStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	stats, err := h.manager.GetInstanceStats(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) HandleInstanceHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	history, err := h.manager.GetInstanceHistory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *Handler) HandleInstanceLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	logPath, err := h.manager.GetInstanceLogPath(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var file *os.File
	for i := 0; i < 10; i++ {
		file, err = os.Open(logPath)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log file not found"})
		return
	}
	defer file.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	reader := bufio.NewReader(file)
	
	c.Stream(func(w io.Writer) bool {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)
				return true
			}
			return false
		}
		c.SSEvent("log", line)
		return true
	})
}

func (h *Handler) HandleClearInstanceLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.manager.ClearInstanceLogs(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logs cleared"})
}

func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "online",
		"region": h.config.Region,
		"uptime": "OK", 
	})
}

func (h *Handler) HandleGetLogs(c *gin.Context) {
	content, err := os.ReadFile("spawner.log")
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"logs": ""})
			return
		}
		h.logger.Error("Failed to read logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": string(content)})
}

func (h *Handler) HandleClearLogs(c *gin.Context) {
	if err := os.Truncate("spawner.log", 0); err != nil {
		h.logger.Error("Failed to clear logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear logs"})
		return
	}
	h.logger.Info("Logs cleared by request")
	c.JSON(http.StatusOK, gin.H{"message": "logs cleared"})
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

func (h *Handler) HandleBackupInstance(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
        return
    }

    if err := h.manager.BackupInstance(id); err != nil {
        h.logger.Error("Failed to backup instance", "id", id, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "backup created"})
}

func (h *Handler) HandleRestoreInstance(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
        return
    }

    var req struct {
        Filename string `json:"filename"`
    }
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if req.Filename == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
        return
    }

    if err := h.manager.RestoreInstance(id, req.Filename); err != nil {
        h.logger.Error("Failed to restore instance", "id", id, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "instance restored"})
}

func (h *Handler) HandleListBackups(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
        return
    }

    backups, err := h.manager.ListBackups(id)
    if err != nil {
        h.logger.Error("Failed to list backups", "id", id, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"backups": backups})
}

func (h *Handler) HandleDeleteBackup(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
        return
    }

    var req struct {
        Filename string `json:"filename"`
    }
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if req.Filename == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
        return
    }

    if err := h.manager.DeleteBackup(id, req.Filename); err != nil {
        h.logger.Error("Failed to delete backup", "id", id, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "backup deleted"})
}