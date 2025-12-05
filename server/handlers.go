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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.Host == "" || s.Port <= 0 || s.Port > 65535 {
		writeError(w, http.StatusBadRequest, "invalid host or port")
		return
	}
	if s.MaxPlayers < 1 || s.MaxPlayers > 10000 {
		writeError(w, http.StatusBadRequest, "invalid max_players")
		return
	}
	id := registry.Register(&s)
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// HeartbeatServer refreshes the LastSeen timestamp for the given server
// ID. This is typically invoked by the game server itself on a periodic
// interval to indicate that it is still alive.
func HeartbeatServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := registry.UpdateHeartbeat(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListServers returns a JSON array with all currently registered servers.
// The response format matches the Server struct JSON representation.
func ListServers(w http.ResponseWriter, r *http.Request) {
	servers := registry.List()
	writeJSON(w, http.StatusOK, servers)
}

// GetServer returns a single server by numeric id. When a server with the
// provided id is not found, the handler responds with 404 Not Found.
func GetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s, ok := registry.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// DeleteServer removes a server from the registry. Only the numeric id is
// required in the URL. Returns 200 OK when deletion succeeds or 404 when
// the server does not exist.
func DeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !registry.Delete(id) {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// Health is a lightweight liveness endpoint used by load balancers and
// orchestrators to confirm the service is running.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
