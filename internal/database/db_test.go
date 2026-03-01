package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB_CreatesCustomPathAndSchema(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "nested", "clippy.db")

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to exist at %s: %v", dbPath, err)
	}

	var tableName string
	err := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='snippets'").Scan(&tableName)
	if err != nil {
		t.Fatalf("expected snippets table to exist: %v", err)
	}
	if tableName != "snippets" {
		t.Fatalf("unexpected table name: %s", tableName)
	}
}
