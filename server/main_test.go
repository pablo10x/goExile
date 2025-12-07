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
		items:  make(map[int]*Spawner),
	}
}

func TestRegisterAndGetSpawner(t *testing.T) {
	resetRegistry()
	body := map[string]interface{}{
		"region":        "room-1",
		"host":          "127.0.0.1",
		"port":          7777,
		"max_instances": 8,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/spawners", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterSpawner(w, req)
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
	req2 := httptest.NewRequest("GET", "/api/spawners/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	GetSpawner(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
	var s Spawner
	if err := json.NewDecoder(w2.Body).Decode(&s); err != nil {
		t.Fatalf("decode spawner: %v", err)
	}
	if s.ID != id || s.Host != "127.0.0.1" || s.Port != 7777 {
		t.Fatalf("unexpected spawner data: %+v", s)
	}
}

func TestHeartbeatUpdatesLastSeen(t *testing.T) {
	resetRegistry()
	id := registry.Register(&Spawner{Region: "hb", Host: "127.0.0.1", Port: 7777, MaxInstances: 4, CurrentInstances: 0, Status: "active"})
	old := registry.items[id].LastSeen
	// small sleep to ensure new timestamp is different
	time.Sleep(10 * time.Millisecond)

	hbBody := map[string]interface{}{
		"current_instances": 1,
		"max_instances":     4,
		"status":            "active",
	}
	b, _ := json.Marshal(hbBody)

	req := httptest.NewRequest("POST", "/api/spawners/"+strconv.Itoa(id)+"/heartbeat", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	HeartbeatSpawner(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", w.Code, w.Body.String())
	}
	if !registry.items[id].LastSeen.After(old) {
		t.Fatalf("expected LastSeen to be updated: old=%v new=%v", old, registry.items[id].LastSeen)
	}
}

func TestListSpawners(t *testing.T) {
	resetRegistry()
	registry.Register(&Spawner{Region: "a", Host: "1.1.1.1", Port: 1111, MaxInstances: 2})
	registry.Register(&Spawner{Region: "b", Host: "2.2.2.2", Port: 2222, MaxInstances: 4})

	req := httptest.NewRequest("GET", "/api/spawners", nil)
	w := httptest.NewRecorder()

	ListSpawners(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var list []Spawner
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 spawners, got %d", len(list))
	}
}

func TestDeleteSpawner(t *testing.T) {
	resetRegistry()
	id := registry.Register(&Spawner{Region: "to-delete", Host: "3.3.3.3", Port: 3333, MaxInstances: 4})

	req := httptest.NewRequest("DELETE", "/api/spawners/"+strconv.Itoa(id), nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	DeleteSpawner(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// now GetSpawner should return 404
	req2 := httptest.NewRequest("GET", "/api/spawners/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	GetSpawner(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found after delete, got %d", w2.Code)
	}
}
