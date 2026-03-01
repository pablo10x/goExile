package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSessionID(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		cookie   *http.Cookie
		query    map[string]string
		expected string
	}{
		{
			name:     "Cookie",
			cookie:   &http.Cookie{Name: "session", Value: "cookie-session"},
			expected: "cookie-session",
		},
		{
			name:     "Bearer Token",
			header:   "Bearer token-session",
			expected: "token-session",
		},
		{
			name:     "Query Token",
			query:    map[string]string{"token": "query-token"},
			expected: "query-token",
		},
		{
			name:     "Query Session",
			query:    map[string]string{"session": "query-session"},
			expected: "query-session",
		},
		{
			name:     "Priority: Cookie over Bearer",
			cookie:   &http.Cookie{Name: "session", Value: "cookie-val"},
			header:   "Bearer token-val",
			expected: "cookie-val",
		},
		{
			name:     "Priority: Bearer over Query",
			header:   "Bearer token-val",
			query:    map[string]string{"token": "query-val"},
			expected: "token-val",
		},
		{
			name:     "No Session",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			q := req.URL.Query()
			for k, v := range tt.query {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			if got := GetSessionID(req); got != tt.expected {
				t.Errorf("GetSessionID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAuthMiddleware_Bearer(t *testing.T) {
	ss := NewSessionStore(false)
	sid, _ := ss.CreateSession(AuthStepAuthenticated)
	cfg := AuthConfig{Enabled: true}

	mw := AuthMiddleware(cfg, ss)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Valid Bearer Token
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+sid)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("AuthMiddleware with valid Bearer token returned %d, want 200", w.Code)
	}

	// Invalid Bearer Token
	req = httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware with invalid Bearer token returned %d, want 401", w.Code)
	}
}
