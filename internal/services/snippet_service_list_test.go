package services

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"path/filepath"
	"testing"
	"time"
)

func TestList_ReturnsAllSnippets(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "list_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets
	snippet1, _ := models.NewSnippet("First Snippet", "code1", "go", []string{"test", "unit"})
	snippet2, _ := models.NewSnippet("Second Snippet", "code2", "python", []string{"python"})
	snippet3, _ := models.NewSnippet("Third Snippet", "code3", "bash", []string{"bash", "script"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet1: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet2: %v", err)
	}
	if _, err := service.Save(snippet3); err != nil {
		t.Fatalf("Failed to save snippet3: %v", err)
	}

	// List all snippets
	snippets, err := service.List(nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snippets) != 3 {
		t.Errorf("Expected 3 snippets, got %d", len(snippets))
	}

	// Should be ordered by created_at descending (newest first)
	if snippets[0].Title != "Third Snippet" {
		t.Errorf("Expected first snippet to be 'Third Snippet', got '%s'", snippets[0].Title)
	}
	if snippets[2].Title != "First Snippet" {
		t.Errorf("Expected last snippet to be 'First Snippet', got '%s'", snippets[2].Title)
	}
}

func TestList_WithTagFilter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "filter_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets with different tags
	snippet1, _ := models.NewSnippet("Docker Cmd", "docker ps", "bash", []string{"docker", "devops"})
	snippet2, _ := models.NewSnippet("Git Cmd", "git log", "bash", []string{"git", "version-control"})
	snippet3, _ := models.NewSnippet("Kubectl Cmd", "kubectl get pods", "bash", []string{"kubernetes", "devops"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet1: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet2: %v", err)
	}
	if _, err := service.Save(snippet3); err != nil {
		t.Fatalf("Failed to save snippet3: %v", err)
	}

	// Filter by "devops" - should return 2 snippets
	snippets, err := service.List([]string{"devops"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(snippets) != 2 {
		t.Errorf("Expected 2 snippets with 'devops' tag, got %d", len(snippets))
	}

	// Check that both have devops tag
	for _, s := range snippets {
		if !s.HasTag("devops") {
			t.Errorf("Snippet '%s' doesn't have 'devops' tag", s.Title)
		}
	}
}

func TestList_WithMultipleTagFilters(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "multi_filter_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets
	snippet1, _ := models.NewSnippet("Go HTTP", "http.Listen", "go", []string{"go", "http"})
	snippet2, _ := models.NewSnippet("Python HTTP", "flask run", "python", []string{"python", "http"})
	snippet3, _ := models.NewSnippet("Bash Script", "echo hi", "bash", []string{"bash"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet1: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet2: %v", err)
	}
	if _, err := service.Save(snippet3); err != nil {
		t.Fatalf("Failed to save snippet3: %v", err)
	}

	// Filter by "go" OR "python" - should return 2 snippets
	snippets, err := service.List([]string{"go", "python"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(snippets) != 2 {
		t.Errorf("Expected 2 snippets with 'go' or 'python' tag, got %d", len(snippets))
	}
}

func TestList_EmptyDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "empty_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	snippets, err := service.List(nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snippets) != 0 {
		t.Errorf("Expected 0 snippets from empty database, got %d", len(snippets))
	}
}

func TestList_FilterReturnsEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "no_match_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create snippets without the filter tag
	snippet1, _ := models.NewSnippet("Test", "code", "go", []string{"test"})
	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Filter by non-existent tag
	snippets, err := service.List([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(snippets) != 0 {
		t.Errorf("Expected 0 snippets with non-existent tag, got %d", len(snippets))
	}
}

func TestList_ParsesTimestampsCorrectly(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "timestamp_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	snippet, _ := models.NewSnippet("Timestamp Test", "code", "go", []string{"test"})
	if _, err := service.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	snippets, err := service.List(nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(snippets) != 1 {
		t.Fatalf("Expected 1 snippet, got %d", len(snippets))
	}

	// Check that timestamp is reasonably recent (within last minute)
	if time.Since(snippets[0].CreatedAt) > time.Minute {
		t.Errorf("Created timestamp seems too old: %v", snippets[0].CreatedAt)
	}
}

func TestList_CaseInsensitiveTagMatching(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "case_test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create snippet with lowercase tags
	snippet, _ := models.NewSnippet("Test", "code", "go", []string{"docker", "python"})
	if _, err := service.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Filter with uppercase tag name
	snippets, err := service.List([]string{"DOCKER"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if len(snippets) != 1 {
		t.Errorf("Expected 1 snippet with 'DOCKER' filter, got %d", len(snippets))
	}
}
