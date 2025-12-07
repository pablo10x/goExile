package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// RegisterSpawner accepts registration from a Spawner service.
func RegisterSpawner(w http.ResponseWriter, r *http.Request) {
	var s Spawner
	if err := decodeJSON(r, &s); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if s.Host == "" {
		writeError(w, r, http.StatusBadRequest, "invalid host")
		return
	}
	if s.Port < 1 || s.Port > 65535 {
		writeError(w, r, http.StatusBadRequest, "invalid port")
		return
	}
	if s.MaxInstances < 0 {
		writeError(w, r, http.StatusBadRequest, "invalid max_instances")
		return
	}

	id := registry.Register(&s)
	GlobalStats.UpdateActiveServers(len(registry.List()))
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatSpawner refreshes the LastSeen timestamp and stats for the given spawner ID.
func HeartbeatSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		CurrentInstances int    `json:"current_instances"`
		MaxInstances     int    `json:"max_instances"`
		Status           string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := registry.UpdateHeartbeat(id, req.CurrentInstances, req.MaxInstances, req.Status); err != nil {
		writeError(w, r, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListSpawners returns a JSON array with all currently registered spawners.
func ListSpawners(w http.ResponseWriter, r *http.Request) {
	spawners := registry.List()
	GlobalStats.UpdateActiveServers(len(spawners))
	writeJSON(w, http.StatusOK, spawners)
}

// GetSpawner returns a single spawner by numeric id.
func GetSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// DeleteSpawner removes a spawner from the registry.
func DeleteSpawner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.Delete(id) {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}
	GlobalStats.UpdateActiveServers(len(registry.List()))
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// SpawnInstance triggers a new game instance on the specified spawner.
func SpawnInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		fmt.Println("error parsing id")
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/spawn", s.Host, s.Port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner: %v", err))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// Health is a lightweight liveness endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// ServeGameServerFile serves the game_server.zip file to authenticated spawners.
func ServeGameServerFile(w http.ResponseWriter, r *http.Request) {
	// In a real app, verify file existence first to give a better error
	http.ServeFile(w, r, "./files/game_server.zip")
}

// StatsAPI returns JSON statistics about the server for the dashboard.
func StatsAPI(w http.ResponseWriter, r *http.Request) {
	totalReq, totalErr, active, dbOK, uptime, mem, tx, rx := GlobalStats.GetStats()

	stats := map[string]interface{}{
		"uptime":          uptime.Milliseconds(),
		"active_spawners": active,
		"total_requests":  totalReq,
		"total_errors":    totalErr,
		"db_connected":    dbOK,
		"memory_usage":    mem,
		"bytes_sent":      tx,
		"bytes_received":  rx,
	}

	writeJSON(w, http.StatusOK, stats)
}

// ErrorsAPI returns the recent error log.
func ErrorsAPI(w http.ResponseWriter, r *http.Request) {
	errors := GlobalStats.GetErrors()
	writeJSON(w, http.StatusOK, errors)
}

// DashboardPage serves the HTML dashboard.
func DashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "webpage/dashboard.html")
}

// ErrorsPage serves the HTML errors page.
func ErrorsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "webpage/errors.html")
}

// LoginPage serves the login form.
func LoginPage(w http.ResponseWriter, r *http.Request, authConfig AuthConfig) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Spawner Registry - Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px; }
        .login-container { width: 100%; max-width: 450px; }
        .login-card { background: white; border-radius: 15px; padding: 50px 40px; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3); }
        .login-header { text-align: center; margin-bottom: 40px; }
        .login-title { font-size: 2em; color: #333; margin-bottom: 10px; font-weight: bold; }
        .login-subtitle { font-size: 0.95em; color: #666; }
        .form-group { margin-bottom: 25px; }
        .form-label { display: block; font-size: 0.9em; font-weight: 600; color: #333; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px; }
        .form-input { width: 100%; padding: 12px 15px; border: 2px solid #e0e0e0; border-radius: 8px; font-size: 1em; transition: all 0.3s ease; font-family: inherit; }
        .form-input:focus { outline: none; border-color: #667eea; box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1); }
        .login-btn { width: 100%; padding: 14px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; border-radius: 8px; font-size: 1em; font-weight: 600; cursor: pointer; transition: all 0.3s ease; text-transform: uppercase; letter-spacing: 1px; margin-top: 30px; }
        .login-btn:hover { transform: translateY(-2px); box-shadow: 0 10px 25px rgba(102, 126, 234, 0.4); }
        .info-box { background-color: #e7f3ff; color: #004085; padding: 15px; border-radius: 8px; margin-bottom: 25px; border-left: 4px solid #0066cc; font-size: 0.85em; }
        .info-box strong { display: block; margin-bottom: 8px; }
        .info-item { margin: 5px 0; font-family: 'Courier New', monospace; }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-card">
            <div class="login-header">
                <div class="login-title">🎮 Spawner Registry</div>
                <div class="login-subtitle">Admin Dashboard</div>
            </div>

            <div class="info-box">
                <strong>📱 Development Mode - Default Credentials</strong>
                <div class="info-item">📧 Email: ` + authConfig.Email + `</div>
                <div class="info-item">🔐 Password: ` + authConfig.Password + `</div>
            </div>

            <form method="POST" action="/login">
                <div class="form-group">
                    <label class="form-label">Email Address</label>
                    <input type="email" name="email" class="form-input" placeholder="admin@example.com" required>
                </div>

                <div class="form-group">
                    <label class="form-label">Password</label>
                    <input type="password" name="password" class="form-input" placeholder="••••••••" required>
                </div>

                <button type="submit" class="login-btn">🔓 Login</button>
            </form>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// HandleLogin processes login requests.
func HandleLogin(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method == http.MethodGet {
		LoginPage(w, r, authConfig)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == authConfig.Email && password == authConfig.Password {
		sessionID, err := sessionStore.CreateSession()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleLogout revokes the session.
func HandleLogout(w http.ResponseWriter, r *http.Request, sessionStore *SessionStore) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionStore.RevokeSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
