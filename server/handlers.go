package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	loginAttempts   = make(map[string]time.Time)
	loginAttemptsMu sync.Mutex
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
		CurrentInstances int     `json:"current_instances"`
		MaxInstances     int     `json:"max_instances"`
		Status           string  `json:"status"`
		CpuUsage         float64 `json:"cpu_usage"`
		MemUsed          uint64  `json:"mem_used"`
		MemTotal         uint64  `json:"mem_total"`
		DiskUsed         uint64  `json:"disk_used"`
		DiskTotal        uint64  `json:"disk_total"`
		GameVersion      string  `json:"game_version"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := registry.UpdateHeartbeat(id, req.CurrentInstances, req.MaxInstances, req.Status, req.CpuUsage, req.MemUsed, req.MemTotal, req.DiskUsed, req.DiskTotal, req.GameVersion); err != nil {
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

	// Increased timeout to 30s to prevent context deadline exceeded on slower networks
	client := &http.Client{Timeout: 30 * time.Second}
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

// ClearSpawnerLogs truncates the log file on a spawner.
func ClearSpawnerLogs(w http.ResponseWriter, r *http.Request) {
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
	req, err := http.NewRequest("DELETE", url, nil)
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

// ListSpawnerInstances fetches active game instances from a spawner.
func ListSpawnerInstances(w http.ResponseWriter, r *http.Request) {
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

	url := fmt.Sprintf("http://%s:%d/instances", s.Host, s.Port)
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

// UpdateSpawnerTemplate triggers the spawner to re-download the latest game server files from master.
func UpdateSpawnerTemplate(w http.ResponseWriter, r *http.Request) {
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

	url := fmt.Sprintf("http://%s:%d/update-template", s.Host, s.Port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}

	client := &http.Client{Timeout: 300 * time.Second} // Long timeout for download
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

// UpdateSpawnerInstance triggers an update (reinstall files) for a specific game instance.
func UpdateSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/update", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}

	client := &http.Client{Timeout: 300 * time.Second} // Long timeout for file copy/download
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

// RenameSpawnerInstance renames a specific game instance on a spawner.
func RenameSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/rename", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("POST", url, r.Body) // Forward body
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

// RemoveSpawnerInstance removes a specific game instance on a spawner.
func RemoveSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("DELETE", url, nil)
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

// StopSpawnerInstance stops a specific game instance on a spawner.
func StopSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/stop", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("POST", url, nil)
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

// StartSpawnerInstance starts a specific game instance on a spawner.

// GetInstanceLogs proxies the log stream from a specific game instance.
func GetInstanceLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/logs", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}

	client := &http.Client{} // Default client, no timeout for streaming?
	// Actually, we need to be careful with timeouts for streams.
	// But standard http.Client has no default timeout.
	
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, fmt.Sprintf("failed to contact spawner: %v", err))
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Stream body
	flusher, ok := w.(http.Flusher)
	if !ok {
		// Fallback for non-flusher
		io.Copy(w, resp.Body)
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}

// ClearInstanceLogs proxies the request to clear an instance's logs.
func ClearInstanceLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/logs", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("DELETE", url, nil)
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

// GetInstanceStats proxies the stats request for a specific game instance.
func GetInstanceStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/stats", s.Host, s.Port, instanceID)
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

// StartSpawnerInstance starts a specific game instance on a spawner.
func StartSpawnerInstance(w http.ResponseWriter, r *http.Request) {
	log.Printf("StartSpawnerInstance handler invoked for request: %s", r.URL.Path)
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	instanceID := vars["instance_id"]
	if instanceID == "" {
		writeError(w, r, http.StatusBadRequest, "missing instance_id")
		return
	}

	s, ok := registry.Get(id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "spawner not found")
		return
	}

	url := fmt.Sprintf("http://%s:%d/instance/%s/start", s.Host, s.Port, instanceID)
	req, err := http.NewRequest("POST", url, nil) // Changed to POST method
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create request")
		return
	}
	// Add content type if the spawner expects a body, even an empty one.
	// For now, assuming no specific body is needed for a simple start.
	// req.Header.Set("Content-Type", "application/json") 


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

	// Rate Limiting: 1 attempt per 30 minutes per IP
	ip := r.RemoteAddr // In prod, consider X-Forwarded-For if behind proxy
	loginAttemptsMu.Lock()
	lastAttempt, exists := loginAttempts[ip]
	if exists && time.Since(lastAttempt) < 30*time.Minute {
		loginAttemptsMu.Unlock()
		// Silent failure: Redirect back to login without error message
		log.Printf("Login rate limited for IP: %s", ip)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	loginAttempts[ip] = time.Now()
	loginAttemptsMu.Unlock()

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
