package services

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearch_EmptyQueryReturnsError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_empty_query.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Empty query should return error
	_, err := service.Search("", "")
	if err == nil {
		t.Error("Expected error for empty query")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestSearch_WhitespaceOnlyQueryReturnsError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_whitespace_query.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Whitespace-only query should return error
	_, err := service.Search("   ", "")
	if err == nil {
		t.Error("Expected error for whitespace-only query")
	}
}

func TestSearch_MatchesTitle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_title.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets
	snippet1, _ := models.NewSnippet("Docker Cleanup Command", "docker ps", "bash", []string{"docker"})
	snippet2, _ := models.NewSnippet("Git Log Command", "git log", "bash", []string{"git"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search for "docker" should match first snippet by title
	results, err := service.Search("docker", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Docker Cleanup Command" {
		t.Errorf("Expected 'Docker Cleanup Command', got '%s'", results[0].Title)
	}
}

func TestSearch_MatchesTags(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_tags.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets
	snippet1, _ := models.NewSnippet("Server Setup", "code", "go", []string{"web", "http"})
	snippet2, _ := models.NewSnippet("Client Code", "code", "python", []string{"cli"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search for "http" should match first snippet by tag
	results, err := service.Search("http", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if !results[0].HasTag("http") {
		t.Error("Expected result to have 'http' tag")
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_case.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippet with uppercase tags
	snippet, _ := models.NewSnippet("DOCKER Command", "docker ps", "bash", []string{"Docker", "DevOps"})
	if _, err := service.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search with lowercase should still match
	results, err := service.Search("docker", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSearch_PartialMatch(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_partial.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets
	snippet1, _ := models.NewSnippet("HTTP Server Example", "code", "go", []string{"web"})
	snippet2, _ := models.NewSnippet("Database Connection", "code", "python", []string{"db"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search for "serve" should partially match "Server" in title
	results, err := service.Search("serve", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Title != "HTTP Server Example" {
		t.Errorf("Expected 'HTTP Server Example', got '%s'", results[0].Title)
	}
}

func TestSearch_WithLanguageFilter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_lang.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create test snippets with same tag but different languages
	snippet1, _ := models.NewSnippet("HTTP Server", "code", "go", []string{"http"})
	snippet2, _ := models.NewSnippet("HTTP Client", "code", "python", []string{"http"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet1: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet2: %v", err)
	}

	// Search for "http" with language filter "go"
	results, err := service.Search("http", "go")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", results[0].Language)
	}
}

func TestSearch_LanguageFilterCaseInsensitive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_lang_case.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	snippet, _ := models.NewSnippet("Test", "code", "PYTHON", []string{"test"})
	if _, err := service.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search with lowercase language filter
	results, err := service.Search("test", "python")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSearch_NoMatchReturnsEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_no_match.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	snippet, _ := models.NewSnippet("Test Snippet", "code", "go", []string{"test"})
	if _, err := service.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	// Search for non-existent term
	results, err := service.Search("nonexistent", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestSearch_MultipleMatchesOrderedByNewest(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "search_multi.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.CloseDB()

	service := NewSnippetService()

	// Create multiple snippets with "http" in title/tags
	snippet1, _ := models.NewSnippet("HTTP 1", "code1", "go", []string{"http"})
	snippet2, _ := models.NewSnippet("HTTP 2", "code2", "python", []string{"http"})
	snippet3, _ := models.NewSnippet("HTTP 3", "code3", "bash", []string{"http"})

	if _, err := service.Save(snippet1); err != nil {
		t.Fatalf("Failed to save snippet1: %v", err)
	}
	if _, err := service.Save(snippet2); err != nil {
		t.Fatalf("Failed to save snippet2: %v", err)
	}
	if _, err := service.Save(snippet3); err != nil {
		t.Fatalf("Failed to save snippet3: %v", err)
	}

	// Search for "http"
	results, err := service.Search("http", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Should be ordered by created_at descending (newest first)
	if results[0].Title != "HTTP 3" {
		t.Errorf("Expected first result to be 'HTTP 3', got '%s'", results[0].Title)
	}
	if results[2].Title != "HTTP 1" {
		t.Errorf("Expected last result to be 'HTTP 1', got '%s'", results[2].Title)
	}
}
