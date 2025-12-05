package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterServer accepts registration from a Unity game server instance.
//
// Expected JSON payload: {"name":"...","host":"...","port":7777,"max_players":8,...}
// On success the handler returns 201 Created with {"id": <server id>}.
func RegisterServer(w http.ResponseWriter, r *http.Request) {
	var s Server
	if err := decodeJSON(r, &s); err != nil {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.Host == "" || s.Port <= 0 || s.Port > 65535 {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, "invalid host or port")
		return
	}
	if s.MaxPlayers < 1 || s.MaxPlayers > 10000 {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, "invalid max_players")
		return
	}
	id := registry.Register(&s)
	GlobalStats.UpdateActiveServers(len(registry.List()))
	GlobalStats.RecordRequest(http.StatusCreated)
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatServer refreshes the LastSeen timestamp for the given server
// ID. This is typically invoked by the game server itself on a periodic
// interval to indicate that it is still alive.
func HeartbeatServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := registry.UpdateHeartbeat(id); err != nil {
		GlobalStats.RecordRequest(http.StatusNotFound)
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListServers returns a JSON array with all currently registered servers.
// The response format matches the Server struct JSON representation.
func ListServers(w http.ResponseWriter, r *http.Request) {
	servers := registry.List()
	GlobalStats.UpdateActiveServers(len(servers))
	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, servers)
}

// GetServer returns a single server by numeric id. When a server with the
// provided id is not found, the handler responds with 404 Not Found.
func GetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.Get(id)
	if !ok {
		GlobalStats.RecordRequest(http.StatusNotFound)
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, s)
}

// DeleteServer removes a server from the registry. Only the numeric id is
// required in the URL. Returns 200 OK when deletion succeeds or 404 when
// the server does not exist.
func DeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		GlobalStats.RecordRequest(http.StatusBadRequest)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.Delete(id) {
		GlobalStats.RecordRequest(http.StatusNotFound)
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	GlobalStats.UpdateActiveServers(len(registry.List()))
	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// Health is a lightweight liveness endpoint used by load balancers and
// orchestrators to confirm the service is running.
func Health(w http.ResponseWriter, r *http.Request) {
	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// StatsAPI returns JSON statistics about the server for the dashboard.
func StatsAPI(w http.ResponseWriter, r *http.Request) {
	totalReq, totalErr, active, dbOK, uptime := GlobalStats.GetStats()

	stats := map[string]interface{}{
		"uptime":         uptime.Milliseconds(),
		"active_servers": active,
		"total_requests": totalReq,
		"total_errors":   totalErr,
		"db_connected":   dbOK,
	}

	GlobalStats.RecordRequest(http.StatusOK)
	writeJSON(w, http.StatusOK, stats)
}

// DashboardPage serves the HTML dashboard.
func DashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Dashboard HTML is embedded or served from file
	// For now, serve the static file
	http.ServeFile(w, r, "dashboard.html")
}

// LoginPage serves the login form.
func LoginPage(w http.ResponseWriter, r *http.Request, authConfig AuthConfig) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Simple HTML login page with credentials displayed in dev mode
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Game Server Registry - Login</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }

        .login-container {
            width: 100%;
            max-width: 450px;
        }

        .login-card {
            background: white;
            border-radius: 15px;
            padding: 50px 40px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }

        .login-header {
            text-align: center;
            margin-bottom: 40px;
        }

        .login-title {
            font-size: 2em;
            color: #333;
            margin-bottom: 10px;
            font-weight: bold;
        }

        .login-subtitle {
            font-size: 0.95em;
            color: #666;
        }

        .form-group {
            margin-bottom: 25px;
        }

        .form-label {
            display: block;
            font-size: 0.9em;
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .form-input {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            transition: all 0.3s ease;
            font-family: inherit;
        }

        .form-input:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .login-btn {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1em;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-top: 30px;
        }

        .login-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 25px rgba(102, 126, 234, 0.4);
        }

        .error-message {
            background-color: #f8d7da;
            color: #721c24;
            padding: 12px 15px;
            border-radius: 8px;
            margin-bottom: 25px;
            border-left: 4px solid #dc3545;
            font-size: 0.9em;
        }

        .info-box {
            background-color: #e7f3ff;
            color: #004085;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 25px;
            border-left: 4px solid #0066cc;
            font-size: 0.85em;
        }

        .info-box strong {
            display: block;
            margin-bottom: 8px;
        }

        .info-item {
            margin: 5px 0;
            font-family: 'Courier New', monospace;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-card">
            <div class="login-header">
                <div class="login-title">🎮 Game Server Registry</div>
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

	// Parse form data
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate credentials
	if email == authConfig.Email && password == authConfig.Password {
		// Create session
		sessionID, err := sessionStore.CreateSession()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400, // 24 hours
		})

		// Redirect to dashboard
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Invalid credentials - show login page with error
	// For now, just redirect back to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleLogout revokes the session.
func HandleLogout(w http.ResponseWriter, r *http.Request, sessionStore *SessionStore) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionStore.RevokeSession(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
