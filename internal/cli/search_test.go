package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSearch_EmptyQueryReturnsError(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "test.db"),
	}

	err := runSearch(opts, "", "")
	if err == nil {
		t.Error("Expected error for empty query")
	}
	if !strings.Contains(err.Error(), "search query cannot be empty") {
		t.Errorf("Expected 'search query cannot be empty' error, got: %v", err)
	}
}

func TestRunSearch_WhitespaceQueryReturnsError(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "test.db"),
	}

	err := runSearch(opts, "   ", "")
	if err == nil {
		t.Error("Expected error for whitespace query")
	}
}

func TestRunSearch_ValidQuerySucceeds(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "test.db"),
	}

	// Should not error even with no results
	err := runSearch(opts, "nonexistent", "")
	if err != nil {
		t.Errorf("Unexpected error for valid search: %v", err)
	}
}

func TestRunSearch_WithLanguageFilter(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "test.db"),
	}

	// Should not error
	err := runSearch(opts, "test", "go")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
