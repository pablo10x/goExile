package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"exile/server/database"
	"exile/server/handlers"
	"exile/server/models"
	"exile/server/registry"

	"github.com/gorilla/mux"
)

func TestMain(m *testing.M) {
	// Initialize in-memory database for tests
	_ = database.InitDB(":memory:")
	os.Exit(m.Run())
}

func resetRegistry() {
	registry.GlobalRegistry.Reset()
	if database.DBConn != nil {
		_, _ = database.DBConn.Exec("DELETE FROM nodes")
	}
}

func TestRegisterAndGetNode(t *testing.T) {
	resetRegistry()
	body := map[string]interface{}{
		"region":        "room-1",
		"host":          "127.0.0.1",
		"port":          7777,
		"max_instances": 8,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/nodes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RegisterNode(w, req)
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
	req2 := httptest.NewRequest("GET", "/api/nodes/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	handlers.GetNode(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
	var s models.Node
	if err := json.NewDecoder(w2.Body).Decode(&s); err != nil {
		t.Fatalf("decode node: %v", err)
	}
	if s.ID != id || s.Host != "127.0.0.1" || s.Port != 7777 {
		t.Fatalf("unexpected node data: %+v", s)
	}
}

func TestHeartbeatUpdatesLastSeen(t *testing.T) {
	resetRegistry()
	id, _ := registry.GlobalRegistry.Register(&models.Node{Region: "hb", Host: "127.0.0.1", Port: 7777, MaxInstances: 4, CurrentInstances: 0, Status: "active"})

	s, _ := registry.GlobalRegistry.Get(id)
	old := s.LastSeen

	// small sleep to ensure new timestamp is different
	time.Sleep(10 * time.Millisecond)

	hbBody := map[string]interface{}{
		"current_instances": 1,
		"max_instances":     4,
		"status":            "active",
	}
	b, _ := json.Marshal(hbBody)

	req := httptest.NewRequest("POST", "/api/nodes/"+strconv.Itoa(id)+"/heartbeat", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	handlers.HeartbeatNode(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", w.Code, w.Body.String())
	}

	s2, _ := registry.GlobalRegistry.Get(id)
	if !s2.LastSeen.After(old) {
		t.Fatalf("expected LastSeen to be updated: old=%v new=%v", old, s2.LastSeen)
	}
}

func TestListNodes(t *testing.T) {
	resetRegistry()
	registry.GlobalRegistry.Register(&models.Node{Region: "a", Host: "1.1.1.1", Port: 1111, MaxInstances: 2})
	registry.GlobalRegistry.Register(&models.Node{Region: "b", Host: "2.2.2.2", Port: 2222, MaxInstances: 4})

	req := httptest.NewRequest("GET", "/api/nodes", nil)
	w := httptest.NewRecorder()

	handlers.ListNodes(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var list []models.Node
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(list))
	}
}

func TestDeleteNode(t *testing.T) {
	resetRegistry()
	id, _ := registry.GlobalRegistry.Register(&models.Node{Region: "to-delete", Host: "3.3.3.3", Port: 3333, MaxInstances: 4})

	req := httptest.NewRequest("DELETE", "/api/nodes/"+strconv.Itoa(id), nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(id)})
	w := httptest.NewRecorder()

	handlers.DeleteNode(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// now GetNode should return 404
	req2 := httptest.NewRequest("GET", "/api/nodes/"+strconv.Itoa(id), nil)
	req2 = mux.SetURLVars(req2, map[string]string{"id": strconv.Itoa(id)})
	w2 := httptest.NewRecorder()
	handlers.GetNode(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found after delete, got %d", w2.Code)
	}
}
