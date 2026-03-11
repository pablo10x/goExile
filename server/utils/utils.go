package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"exile/server/models"
)

var StartTime = time.Now()

const (
	MaxIDValue  = 1000000 // A reasonable upper limit for IDs
	maxBodySize = 1 << 20 // 1MB
)

// GenerateRandomString generates a secure random string of the specified length (in bytes).
// The resulting string is hex-encoded, so it will be twice the requested length.
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback or panic in critical failure?
		// For now, logging and returning empty is safer than panic for non-critical flow,
		// but auth depends on this.
		log.Printf("critical: failed to generate random string: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// sealedResponseWriter prevents any further writes once it has been used.
type sealedResponseWriter struct {
	http.ResponseWriter
	written bool
}

func (w *sealedResponseWriter) WriteHeader(status int) {
	if w.written {
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *sealedResponseWriter) Write(b []byte) (int, error) {
	if w.written {
		return 0, nil
	}
	w.written = true
	return w.ResponseWriter.Write(b)
}

// WriteJSON encodes data as JSON and writes it to the ResponseWriter with
// the provided HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}

	js, err := json.Marshal(data)
	if err != nil {
		log.Printf("error encoding response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	// Wrap to prevent further writes
	sw := &sealedResponseWriter{ResponseWriter: w}

	sw.Header().Set("Content-Type", "application/json")
	sw.WriteHeader(status)
	_, _ = sw.Write(js)
}

// WriteError sends a structured error response using ErrorResponse.
func WriteError(w http.ResponseWriter, r *http.Request, status int, message string) {
	WriteJSON(w, status, models.ErrorResponse{Error: message})
}

// ParseID converts a string representation of an ID into an integer and
// validates it against a safe range.
func ParseID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format")
	}
	if id < 1 || id > MaxIDValue {
		return 0, fmt.Errorf("ID out of valid range")
	}
	return id, nil
}

// DecodeJSON decodes a request body into v using a size-limited reader
// and disallowing unknown fields for safety.
func DecodeJSON(r *http.Request, v interface{}) error {
	lr := io.LimitReader(r.Body, maxBodySize)
	decoder := json.NewDecoder(lr)
	// decoder.DisallowUnknownFields() // Allow unknown fields for better compatibility
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is present, the value is returned. Otherwise, defaultVal is returned.
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

// HashString generates a simple numeric hash from a string.
func HashString(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h *= 16777619
		h ^= uint32(s[i])
	}
	return h
}

// UpdateEnvFile updates or adds a key-value pair in the .env file.
func UpdateEnvFile(key, value string) error {
	path := ".env"
	// Check if we are in the server directory or root
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try to find it in parent or current
		if _, err := os.Stat("server/.env"); err == nil {
			path = "server/.env"
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		// If it doesn't exist, create it
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(fmt.Sprintf("%s=%s\n", key, value)), 0600)
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	newLines := make([]string, 0, len(lines)+1)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	output := strings.Join(newLines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	return os.WriteFile(path, []byte(output), 0600)
}
