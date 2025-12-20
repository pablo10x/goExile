package main_test

import (
	"os"
	"testing"
	"time"

	"exile/server/database"
	"exile/server/models"
)

func TestDBVersions(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set, skipping DB test")
	}

	err := database.InitDB(dsn)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	db := database.DBConn
	defer database.CloseDB(db)

	v1 := &models.GameServerVersion{
		Filename:   "test_v1.zip",
		Comment:    "Initial upload",
		UploadedAt: time.Now().UTC(),
		IsActive:   false,
	}

	// 1. Save v1
	if err := database.SaveServerVersion(db, v1); err != nil {
		t.Fatalf("save v1: %v", err)
	}

	// 2. List
	list, err := database.ListServerVersions(db)
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
	if err := database.SetActiveVersion(db, v1ID); err != nil {
		t.Fatalf("set active: %v", err)
	}

	// 4. Get Active
	active, err := database.GetActiveServerVersion(db)
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
	v2 := &models.GameServerVersion{
		Filename:   "test_v2.zip",
		Comment:    "Second upload",
		UploadedAt: time.Now().UTC(),
		IsActive:   false,
	}
	if err := database.SaveServerVersion(db, v2); err != nil {
		t.Fatalf("save v2: %v", err)
	}

	// 6. Switch Active to v2
	// Retrieve v2 ID
	list, _ = database.ListServerVersions(db)
	var v2ID int
	for _, v := range list {
		if v.Filename == "test_v2.zip" {
			v2ID = v.ID
			break
		}
	}

	if err := database.SetActiveVersion(db, v2ID); err != nil {
		t.Fatalf("switch active to v2: %v", err)
	}

	active, err = database.GetActiveServerVersion(db)
	if err != nil {
		t.Fatalf("get active 2: %v", err)
	}
	if active.ID != v2ID {
		t.Fatalf("expected active ID %d, got %d", v2ID, active.ID)
	}

	// 7. Delete v1
	filename, err := database.DeleteServerVersion(db, v1ID)
	if err != nil {
		t.Fatalf("delete v1: %v", err)
	}
	if filename != "test_v1.zip" {
		t.Fatalf("expected filename test_v1.zip, got %s", filename)
	}

	list, _ = database.ListServerVersions(db)
	if len(list) != 1 {
		t.Fatalf("expected 1 version after delete, got %d", len(list))
	}
}
