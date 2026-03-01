package services

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"path/filepath"
	"testing"
)

func TestSave_FailsWithoutDBInitialization(t *testing.T) {
	_ = database.CloseDB()
	service := NewSnippetService()

	snippet, err := models.NewSnippet("Test", "echo hi", "bash", []string{"test"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}

	if _, err := service.Save(snippet); err == nil {
		t.Fatal("expected error when database is not initialized")
	}
}

func TestSave_SucceedsWithInitializedDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "service_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()
	snippet, err := models.NewSnippet("Service Save", "echo ok", "bash", []string{"svc"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}

	id, err := service.Save(snippet)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}
