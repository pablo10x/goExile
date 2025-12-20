package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"exile/server/models"
)

const (
	MaxIDValue  = 1000000 // A reasonable upper limit for IDs
	maxBodySize = 1 << 20 // 1MB
)

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
func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
