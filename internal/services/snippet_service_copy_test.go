package services

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"errors"
	"path/filepath"
	"testing"
)

func TestGetById_Success(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create a test snippet
	snippet, err := models.NewSnippet("Test Snippet", "echo test", "bash", []string{"test", "copy"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}

	id, err := service.Save(snippet)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve by ID
	retrieved, err := service.GetById(id)
	if err != nil {
		t.Fatalf("GetById failed: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("expected ID %d, got %d", id, retrieved.ID)
	}
	if retrieved.Title != "Test Snippet" {
		t.Errorf("expected title 'Test Snippet', got '%s'", retrieved.Title)
	}
	if retrieved.Code != "echo test" {
		t.Errorf("expected code 'echo test', got '%s'", retrieved.Code)
	}
	if retrieved.Language != "bash" {
		t.Errorf("expected language 'bash', got '%s'", retrieved.Language)
	}
	if len(retrieved.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(retrieved.Tags))
	}
}

func TestGetById_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Try to get a non-existent ID
	_, err := service.GetById(99999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if !errors.Is(err, ErrSnippetNotFound) {
		t.Errorf("expected ErrSnippetNotFound, got %v", err)
	}
}

func TestGetById_InvalidID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	tests := []struct {
		name    string
		id      int64
		wantErr error
	}{
		{"zero ID", 0, ErrInvalidSnippetID},
		{"negative ID", -1, ErrInvalidSnippetID},
		{"negative large ID", -999, ErrInvalidSnippetID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetById(tt.id)
			if err == nil {
				t.Fatal("expected error for invalid ID")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestGetById_NoDatabase(t *testing.T) {
	// Close any existing database
	database.CloseDB()

	service := NewSnippetService()

	_, err := service.GetById(1)
	if err == nil {
		t.Fatal("expected error when database is not initialized")
	}
}

func TestGetByTitle_Success(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create a test snippet with a unique title
	snippet, err := models.NewSnippet("UniqueTitleForTest", "echo unique", "bash", []string{"unique"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}

	id, err := service.Save(snippet)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve by title
	retrieved, err := service.GetByTitle("UniqueTitleForTest")
	if err != nil {
		t.Fatalf("GetByTitle failed: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("expected ID %d, got %d", id, retrieved.ID)
	}
	if retrieved.Title != "UniqueTitleForTest" {
		t.Errorf("expected title 'UniqueTitleForTest', got '%s'", retrieved.Title)
	}
	if retrieved.Code != "echo unique" {
		t.Errorf("expected code 'echo unique', got '%s'", retrieved.Code)
	}
}

func TestGetByTitle_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Try to get a non-existent title
	_, err := service.GetByTitle("NonExistentTitle")
	if err == nil {
		t.Fatal("expected error for non-existent title")
	}
	if !errors.Is(err, ErrSnippetNotFound) {
		t.Errorf("expected ErrSnippetNotFound, got %v", err)
	}
}

func TestGetByTitle_EmptyTitle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	tests := []struct {
		name  string
		title string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tabs and spaces", "\t  \t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetByTitle(tt.title)
			if err == nil {
				t.Fatal("expected error for empty title")
			}
		})
	}
}

func TestGetByTitle_NoDatabase(t *testing.T) {
	// Close any existing database
	database.CloseDB()

	service := NewSnippetService()

	_, err := service.GetByTitle("test")
	if err == nil {
		t.Fatal("expected error when database is not initialized")
	}
}

func TestGetByTitle_TrimsWhitespace(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create a snippet
	snippet, err := models.NewSnippet("TrimTest", "echo trim", "bash", []string{"trim"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}

	_, err = service.Save(snippet)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Retrieve with whitespace padding - should still work
	retrieved, err := service.GetByTitle("  TrimTest  ")
	if err != nil {
		t.Fatalf("GetByTitle with whitespace failed: %v", err)
	}

	if retrieved.Title != "TrimTest" {
		t.Errorf("expected title 'TrimTest', got '%s'", retrieved.Title)
	}
}
