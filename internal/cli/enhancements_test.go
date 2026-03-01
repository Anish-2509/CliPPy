package cli

import (
	"clippy/internal/database"
	"os"
	"strings"
	"testing"
)

func TestIsInputPiped_WithFile(t *testing.T) {
	// Create a temp file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Redirected file input should be considered a valid stdin input source
	if !isInputPiped(tmpfile) {
		t.Error("Regular redirected file should be considered stdin input")
	}
}

func TestIsInputPiped_WithStringReader(t *testing.T) {
	// strings.Reader is not an *os.File, so should return true (for testing)
	r := strings.NewReader("test content")
	if !isInputPiped(r) {
		t.Error("strings.Reader should be considered piped for testing purposes")
	}
}

func TestIsInputPiped_WithNil(t *testing.T) {
	if isInputPiped(nil) {
		t.Error("nil reader should not be considered piped")
	}
}

func TestResolveCodeSource_PrecedenceWithPipedCheck(t *testing.T) {
	tests := []struct {
		name     string
		codeFlag string
		stdin    *strings.Reader
		want     string
		wantErr  bool
		isPiped  bool // Expected behavior for stdin source
	}{
		{
			name:     "--code flag provided, ignores stdin",
			codeFlag: "code from flag",
			stdin:    strings.NewReader("code from stdin"),
			want:     "code from flag",
			wantErr:  false,
		},
		{
			name:     "--code empty, stdin has content (piped)",
			codeFlag: "",
			stdin:    strings.NewReader("code from stdin"),
			want:     "code from stdin",
			wantErr:  false,
		},
		{
			name:     "--code empty, stdin empty",
			codeFlag: "",
			stdin:    strings.NewReader(""),
			want:     "",
			wantErr:  true,
		},
		{
			name:     "--code whitespace, stdin empty",
			codeFlag: "   ",
			stdin:    strings.NewReader(""),
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCodeSource(tt.stdin, tt.codeFlag)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !strings.Contains(err.Error(), "code content required") {
					t.Errorf("expected 'code content required' error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDatabase_DefaultPath(t *testing.T) {
	path, err := database.GetDefaultDBPath()
	if err != nil {
		t.Errorf("failed to get default DB path: %v", err)
	}
	if path == "" {
		t.Error("default DB path should not be empty")
	}
	if !strings.Contains(path, ".clippy") {
		t.Errorf("default DB path should contain .clippy, got: %s", path)
	}
}

func TestDatabase_InitAndClose(t *testing.T) {
	// Use a temp file for testing
	tmpfile, err := os.CreateTemp("", "clippy_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Initialize DB
	if err := database.InitDB(tmpfile.Name()); err != nil {
		t.Fatalf("failed to init DB: %v", err)
	}

	if !database.IsInitialized() {
		t.Error("database should be initialized after InitDB")
	}

	// Close DB
	if err := database.CloseDB(); err != nil {
		t.Errorf("failed to close DB: %v", err)
	}

	if database.IsInitialized() {
		t.Error("database should not be initialized after CloseDB")
	}
}

func TestDatabase_Reinitialize(t *testing.T) {
	// Use a temp file for testing
	tmpfile, err := os.CreateTemp("", "clippy_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Initialize DB twice - should not error
	if err := database.InitDB(tmpfile.Name()); err != nil {
		t.Fatalf("failed to init DB: %v", err)
	}

	// Second init should be idempotent
	if err := database.InitDB(tmpfile.Name()); err != nil {
		t.Errorf("re-init should not error: %v", err)
	}

	database.CloseDB()
}

func TestDatabase_InitWithEmptyPath(t *testing.T) {
	// Empty path should use default
	if err := database.InitDB(""); err != nil {
		t.Errorf("init with empty path should use default: %v", err)
	}
	if !database.IsInitialized() {
		t.Error("database should be initialized")
	}
	database.CloseDB()
}
