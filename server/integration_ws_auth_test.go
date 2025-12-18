package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

// TestWebSocketAuthentication verifies that the WebSocket endpoint is correctly protected by the UnifiedAuthMiddleware.
func TestWebSocketAuthentication(t *testing.T) {
	// 1. Setup Dependencies
	testAPIKey := "test-secret-key"
	authConfig := AuthConfig{Enabled: true} // Minimal config
	sessionStore := NewSessionStore(false)

	// 2. Setup Middleware and Handler
	// We use a simple handler that attempts to upgrade the connection.
	// If the middleware works, this handler should only be reached when authorized.
	upgrader := websocket.Upgrader{}
	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// This might happen if the test client doesn't actually do a WS handshake,
			// but we will use a proper WS dialer in the test.
			t.Logf("Upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		// If we get here, connection is established.
	})

	// Wrap the handler with the middleware
	protectedHandler := UnifiedAuthMiddleware(testAPIKey, authConfig, sessionStore)(wsHandler)

	// Start a test server
	server := httptest.NewServer(protectedHandler)
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + server.URL[4:]

	// 3. Test Cases

	t.Run("Connect without API Key", func(t *testing.T) {
		// Attempt to connect without headers
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		
		// We expect an error from Dial because the server returns 401, not 101 Switching Protocols
		if err == nil {
			t.Fatal("Expected error connecting without API key, but got success")
		}

		// Check the response status code
		if resp == nil {
			t.Fatal("Expected HTTP response, got nil")
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("Connect with Invalid API Key", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("X-API-Key", "wrong-key")

		_, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)

		if err == nil {
			t.Fatal("Expected error connecting with invalid API key, but got success")
		}

		if resp == nil {
			t.Fatal("Expected HTTP response, got nil")
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("Connect with Valid API Key", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("X-API-Key", testAPIKey)

		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatalf("Failed to connect with valid API key: %v", err)
		}
		defer conn.Close()

		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Errorf("Expected status 101 Switching Protocols, got %d", resp.StatusCode)
		}
	})
}
