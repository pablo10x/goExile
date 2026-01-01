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

// WriteJSON encodes data as JSON and writes it to the ResponseWriter with
// the provided HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
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
	decoder.DisallowUnknownFields()
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
