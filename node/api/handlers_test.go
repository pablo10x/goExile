package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"node/internal/config"
	"node/internal/game"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupHandler() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.Config{
		Region: "TestRegion",
	}
	manager := game.NewManager(cfg, logger)
	handler := NewHandler(manager, cfg, logger)

	r := gin.New()
	handler.RegisterRoutes(r)
	return r, handler
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupHandler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)

	if response["region"] != "TestRegion" {
		t.Errorf("Expected region TestRegion, got %v", response["region"])
	}
}

// No longer need AuthMiddleware tests as it's removed.
// func TestAuthMiddleware_MissingKey(t *testing.T) {
// 	router, _ := setupHandler()

// 	w := httptest.NewRecorder()
// 	// Try to access protected route without key
// 	req, _ := http.NewRequest("GET", "/instances", nil)
// 	router.ServeHTTP(w, req)

// 	if w.Code != http.StatusUnauthorized {
// 		t.Errorf("Expected status 401 Unauthorized, got %d", w.Code)
// 	}
// }

// func TestAuthMiddleware_ValidKey(t *testing.T) {
// 	router, _ := setupHandler()

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/instances", nil)
// 	req.Header.Set("X-API-Key", "secret")
// 	router.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Errorf("Expected status 200, got %d", w.Code)
// 	}
// }
