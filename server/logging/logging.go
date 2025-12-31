package logging

import (
	"log/slog"
	"strings"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/registry"
)

type LogCategory string

const (
	LogCategoryInternal LogCategory = "Internal"
	LogCategoryNode  LogCategory = "Node"
	LogCategorySecurity LogCategory = "Security"
	LogCategoryGeneral  LogCategory = "General"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LoggerService handles structured logging.
type LoggerService struct {
	// We can add deps here if needed
}

var Logger = &LoggerService{}

func (s *LoggerService) Log(level LogLevel, category LogCategory, msg string, details string, rPath, rMethod, clientIP string, statusCode int) {
	l := &models.SystemLog{
		Timestamp: time.Now().UTC(),
		Level:     string(level),
		Category:  string(category),
		Message:   msg,
		Details:   details,
		Path:      rPath,
		Method:    rMethod,
		ClientIP:  clientIP,
		Source:    "backend",
	}

	if database.DBConn != nil {
		go func() {
			if err := database.SaveSystemLog(database.DBConn, l); err != nil {
				slog.Error("Failed to save system log", "error", err)
			}
		}()
	}

	// Update In-Memory Stats for Dashboard
	if level == LogLevelError || level == LogLevelFatal {
		registry.GlobalStats.RecordError(rPath, statusCode, msg, clientIP, string(category))
	}
}

func DetermineCategory(path string) LogCategory {
	if strings.HasPrefix(path, "/api/nodes") {
		return LogCategoryNode
	}
	if strings.Contains(path, "login") || strings.Contains(path, "auth") || strings.Contains(path, "security") {
		return LogCategorySecurity
	}
	if strings.HasPrefix(path, "/api") {
		return LogCategoryInternal
	}
	return LogCategoryGeneral
}
