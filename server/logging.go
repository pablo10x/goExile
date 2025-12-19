package main

import (
	"log/slog"
	"strings"
	"time"
)

type LogCategory string

const (
	LogCategoryInternal LogCategory = "Internal"
	LogCategorySpawner  LogCategory = "Spawner"
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
	l := &SystemLog{
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

	if dbConn != nil {
		go func() {
			if err := SaveSystemLog(dbConn, l); err != nil {
				slog.Error("Failed to save system log", "error", err)
			}
		}()
	}

	// Update In-Memory Stats for Dashboard
	if level == LogLevelError || level == LogLevelFatal {
		GlobalStats.RecordError(rPath, statusCode, msg, clientIP, string(category))
	}
}

func DetermineCategory(path string) LogCategory {
	if strings.HasPrefix(path, "/api/spawners") {
		return LogCategorySpawner
	}
	if strings.Contains(path, "login") || strings.Contains(path, "auth") || strings.Contains(path, "security") {
		return LogCategorySecurity
	}
	if strings.HasPrefix(path, "/api") {
		return LogCategoryInternal
	}
	return LogCategoryGeneral
}
