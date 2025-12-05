package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func resetRegistry() {
	registry = &Registry{
		nextID: 1,
		items:  make(map[int]*Server),
	}
}

func TestRegisterAndGetServer(t *testing.T) {
	resetRegistry()
	body := map[string]interface{}{
		"name":        "room-1",
		"host":        "127.0.0.1",
		"port":        7777,
		"max_players": 8,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/servers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode resp: %v", err)
	}
	id := resp["id"]
	if id == 0 {
		t.Fatalf("expected non-zero id")
	}

	// fetch
	req2 := httptest.NewRequest("GET", "/api/servers/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	GetServer(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
	var s Server
	if err := json.NewDecoder(w2.Body).Decode(&s); err != nil {
		t.Fatalf("decode server: %v", err)
	}
	if s.ID != id || s.Host != "127.0.0.1" || s.Port != 7777 {
		t.Fatalf("unexpected server data: %+v", s)
	}
}

func TestHeartbeatUpdatesLastSeen(t *testing.T) {
	resetRegistry()
	id := registry.Register(&Server{Name: "hb", Host: "127.0.0.1", Port: 7777, MaxPlayers: 4})
	old := registry.items[id].LastSeen
	// small sleep to ensure new timestamp is different
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest("POST", "/api/servers/"+strconv.Itoa(id)+"/heartbeat", nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	HeartbeatServer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if !registry.items[id].LastSeen.After(old) {
		t.Fatalf("expected LastSeen to be updated: old=%v new=%v", old, registry.items[id].LastSeen)
	}
}

func TestListServers(t *testing.T) {
	resetRegistry()
	registry.Register(&Server{Name: "a", Host: "1.1.1.1", Port: 1111, MaxPlayers: 2})
	registry.Register(&Server{Name: "b", Host: "2.2.2.2", Port: 2222, MaxPlayers: 4})

	req := httptest.NewRequest("GET", "/api/servers", nil)
	w := httptest.NewRecorder()

	ListServers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var list []Server
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(list))
	}
}

func TestDeleteServer(t *testing.T) {
	resetRegistry()
	id := registry.Register(&Server{Name: "to-delete", Host: "3.3.3.3", Port: 3333, MaxPlayers: 4})

	req := httptest.NewRequest("DELETE", "/api/servers/"+strconv.Itoa(id), nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	DeleteServer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// now GetServer should return 404
	req2 := httptest.NewRequest("GET", "/api/servers/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	GetServer(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found after delete, got %d", w2.Code)
	}
}
