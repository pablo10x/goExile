package database_test

import (
	"os"
	"testing"
	"time"

	"exile/server/database"
	"exile/server/models"
)

func TestDBSaveLoadDelete(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN not set, skipping DB test")
	}

	err := database.InitDB(dsn)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	db := database.DBConn

	// Use random port to avoid unique constraint violation if cleanup fails
	s := &models.Node{
		Region:           "db1",
		Host:             "127.0.0.1",
		Port:             7000 + int(time.Now().UnixNano()%1000),
		MaxInstances:     10,
		CurrentInstances: 0,
		Status:           "active",
		LastSeen:         time.Now().UTC(),
	}

	id, err := database.SaveNode(db, s)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id from SaveNode")
	}

	got, err := database.GetNodeByID(db, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatalf("expected node, got nil")
	}
	if got.Region != s.Region || got.Host != s.Host || got.Port != s.Port {
		t.Fatalf("mismatch node: got=%+v want=%+v", got, s)
	}

	list, err := database.LoadNodes(db)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected at least 1 node in DB")
	}

	if err := database.DeleteNodeDB(db, id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got2, err := database.GetNodeByID(db, id)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if got2 != nil {
		t.Fatalf("expected nil after delete, got %+v", got2)
	}
}
