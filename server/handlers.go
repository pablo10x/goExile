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
