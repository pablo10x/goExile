package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDBVersions(t *testing.T) {
	f, err := os.CreateTemp("", "version_test_*.db")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	db, err := InitDB(path)
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") || strings.Contains(err.Error(), "CGO_ENABLED") {
			t.Skipf("sqlite3 driver not available: %v", err)
			return
		}
		t.Fatalf("init db: %v", err)
	}
	defer CloseDB(db)

	v1 := &GameServerVersion{
		Filename:   "test_v1.zip",
		Comment:    "Initial upload",
		UploadedAt: time.Now().UTC(),
		IsActive:   false,
	}

	// 1. Save v1
	if err := SaveServerVersion(db, v1); err != nil {
		t.Fatalf("save v1: %v", err)
	}

	// 2. List
	list, err := ListServerVersions(db)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 version, got %d", len(list))
	}
	// ID should be assigned
	v1ID := list[0].ID
	if v1ID == 0 {
		t.Fatalf("expected non-zero ID")
	}

	// 3. Set Active
	if err := SetActiveVersion(db, v1ID); err != nil {
		t.Fatalf("set active: %v", err)
	}

	// 4. Get Active
	active, err := GetActiveServerVersion(db)
	if err != nil {
		t.Fatalf("get active: %v", err)
	}
	if active == nil {
		t.Fatalf("expected active version")
	}
	if active.ID != v1ID {
		t.Fatalf("expected active ID %d, got %d", v1ID, active.ID)
	}

	// 5. Add v2
	v2 := &GameServerVersion{
		Filename:   "test_v2.zip",
		Comment:    "Second upload",
		UploadedAt: time.Now().UTC(),
		IsActive:   false,
	}
	if err := SaveServerVersion(db, v2); err != nil {
		t.Fatalf("save v2: %v", err)
	}

	// 6. Switch Active to v2
	// Retrieve v2 ID
	list, _ = ListServerVersions(db)
	var v2ID int
	for _, v := range list {
		if v.Filename == "test_v2.zip" {
			v2ID = v.ID
			break
		}
	}

	if err := SetActiveVersion(db, v2ID); err != nil {
		t.Fatalf("switch active to v2: %v", err)
	}

	active, err = GetActiveServerVersion(db)
	if err != nil {
		t.Fatalf("get active 2: %v", err)
	}
	if active.ID != v2ID {
		t.Fatalf("expected active ID %d, got %d", v2ID, active.ID)
	}

	// 7. Delete v1
	filename, err := DeleteServerVersion(db, v1ID)
	if err != nil {
		t.Fatalf("delete v1: %v", err)
	}
	if filename != "test_v1.zip" {
		t.Fatalf("expected filename test_v1.zip, got %s", filename)
	}

	list, _ = ListServerVersions(db)
	if len(list) != 1 {
		t.Fatalf("expected 1 version after delete, got %d", len(list))
	}
}
