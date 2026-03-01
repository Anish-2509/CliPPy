package cli

import (
	"path/filepath"
	"testing"
)

func TestRunList_EmptyDatabaseShowsMessage(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "empty.db"),
	}

	err := runList(opts, "")
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
	// Function should complete without error even with empty database
}

func TestRunList_WithTagsFilter(t *testing.T) {
	opts := &CLIOptions{
		DBPath: filepath.Join(t.TempDir(), "test.db"),
	}

	err := runList(opts, "docker,go")
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}
