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

	id, err := registry.Register(&s)
	if err != nil {
		// Log the detailed error for debugging
		fmt.Printf("Registration failed: %v\n", err)
		writeError(w, r, http.StatusInternalServerError, "failed to register spawner")
		return
	}

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
	rawID := vars["id"]
	id, err := parseID(rawID)
	if err != nil {
		fmt.Printf("Error parsing ID. Raw: '%s', Error: %v\n", rawID, err)
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

// GetSpawnerLogs fetches and returns the log file content from a spawner.
func GetSpawnerLogs(w http.ResponseWriter, r *http.Request) {
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

	url := fmt.Sprintf("http://%s:%d/logs", s.Host, s.Port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}

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

// ClearErrorsAPI clears the error log.
func ClearErrorsAPI(w http.ResponseWriter, r *http.Request) {
	GlobalStats.ClearErrors()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Error log cleared"})
}

// HandleLogin processes login requests.
func HandleLogin(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method == http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "GET login not supported. Please POST credentials.")
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
