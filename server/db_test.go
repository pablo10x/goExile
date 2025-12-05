package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDBSaveLoadDelete(t *testing.T) {
	f, err := os.CreateTemp("", "registry_test_*.db")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	db, err := InitDB(path)
	if err != nil {
		// In some environments (CGO disabled) the mattn/go-sqlite3 driver is not usable.
		// Detect that case and skip the DB tests instead of failing the whole suite.
		if strings.Contains(err.Error(), "requires cgo") || strings.Contains(err.Error(), "CGO_ENABLED") {
			t.Skipf("sqlite3 driver not available in this environment: %v", err)
			return
		}
		t.Fatalf("init db: %v", err)
	}
	defer db.Close()

	s := &Server{
		Name:           "db1",
		Host:           "127.0.0.1",
		Port:           7777,
		MaxPlayers:     10,
		CurrentPlayers: 0,
		Region:         "us-west",
		Status:         "active",
		LastSeen:       time.Now().UTC(),
	}

	id, err := SaveServer(db, s)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id from SaveServer")
	}

	got, err := GetServerByID(db, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatalf("expected server, got nil")
	}
	if got.Name != s.Name || got.Host != s.Host || got.Port != s.Port {
		t.Fatalf("mismatch server: got=%+v want=%+v", got, s)
	}

	list, err := LoadServers(db)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected at least 1 server in DB")
	}

	if err := DeleteServerDB(db, id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got2, err := GetServerByID(db, id)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if got2 != nil {
		t.Fatalf("expected nil after delete, got %+v", got2)
	}
}
