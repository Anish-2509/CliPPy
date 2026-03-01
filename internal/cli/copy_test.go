package cli

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"clippy/internal/services"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupCopyTestDB creates a test database with sample snippets
func setupCopyTestDB(t *testing.T, dbPath string) (*services.SnippetService, int64, int64) {
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	service := services.NewSnippetService()

	// Create test snippets
	snippet1, err := models.NewSnippet("Docker Prune", "docker system prune -af", "bash", []string{"docker", "cleanup"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}
	id1, err := service.Save(snippet1)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	snippet2, err := models.NewSnippet("Go HTTP Server", `http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
})`, "go", []string{"http", "server"})
	if err != nil {
		t.Fatalf("unexpected model error: %v", err)
	}
	id2, err := service.Save(snippet2)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	return service, id1, id2
}

func TestRunCopy_ArgValidationMatrix(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		title       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no args and no title",
			args:        []string{},
			title:       "",
			wantErr:     true,
			errContains: "either [id] or --title is required",
		},
		{
			name:        "both id and title provided",
			args:        []string{"1"},
			title:       "Some Title",
			wantErr:     true,
			errContains: "cannot use both [id] and --title together",
		},
		{
			name:        "valid id only",
			args:        []string{"1"},
			title:       "",
			wantErr:     false,
		},
		{
			name:        "valid title only",
			args:        []string{},
			title:       "Docker Prune",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			_, _, _ = setupCopyTestDB(t, dbPath)
			defer database.CloseDB()

			mockClipboard := &services.MockClipboard{}
			var stdout strings.Builder

			opts := &CLIOptions{
				DBPath:    dbPath,
				Stdout:    &stdout,
				Stderr:    &strings.Builder{},
				Clipboard: mockClipboard,
			}

			err := runCopy(opts, tt.args, tt.title, 0)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got none", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunCopy_InvalidIDFormat(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{"non-numeric", []string{"abc"}, "must be a positive integer"},
		{"zero", []string{"0"}, "must be a positive integer"},
		{"negative", []string{"-5"}, "must be a positive integer"},
		{"float", []string{"1.5"}, "must be a positive integer"},
		{"with letters", []string{"1abc"}, "must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each sub-test gets its own database and opts
			dbPath := filepath.Join(t.TempDir(), "test.db")
			_, _, _ = setupCopyTestDB(t, dbPath)
			defer database.CloseDB()

			mockClipboard := &services.MockClipboard{}
			opts := &CLIOptions{
				DBPath:    dbPath,
				Stdout:    &strings.Builder{},
				Stderr:    &strings.Builder{},
				Clipboard: mockClipboard,
			}

			err := runCopy(opts, tt.args, "", 0)
			if err == nil {
				t.Fatal("expected error for invalid ID format")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestRunCopy_SuccessByID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, _ = setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	mockClipboard := &services.MockClipboard{}
	var stdout strings.Builder

	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &stdout,
		Stderr:    &strings.Builder{},
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{"1"}, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify clipboard content
	if mockClipboard.Content != "docker system prune -af" {
		t.Errorf("expected clipboard content 'docker system prune -af', got %q", mockClipboard.Content)
	}

	// Verify success message
	output := stdout.String()
	if !strings.Contains(output, "Copied snippet") {
		t.Errorf("expected success message, got %q", output)
	}
	if !strings.Contains(output, "ID:") {
		t.Errorf("expected ID in output, got %q", output)
	}
	if !strings.Contains(output, "ID: 1") {
		t.Errorf("expected ID 1 in output, got %q", output)
	}
}

func TestRunCopy_SuccessByTitle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	mockClipboard := &services.MockClipboard{}
	var stdout strings.Builder

	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &stdout,
		Stderr:    &strings.Builder{},
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{}, "Docker Prune", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify clipboard content
	if mockClipboard.Content != "docker system prune -af" {
		t.Errorf("expected clipboard content 'docker system prune -af', got %q", mockClipboard.Content)
	}

	// Verify success message contains title
	output := stdout.String()
	if !strings.Contains(output, "Docker Prune") {
		t.Errorf("expected title in output, got %q", output)
	}
}

func TestRunCopy_SnippetNotFoundByID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, _ = setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	mockClipboard := &services.MockClipboard{}
	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &strings.Builder{},
		Stderr:    &strings.Builder{},
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{"99999"}, "", 0)
	if err == nil {
		t.Fatal("expected error for non-existent snippet")
	}
	// Accept "snippet not found" or "snippet is nil" as valid error messages
	if !strings.Contains(err.Error(), "snippet not found") && !strings.Contains(err.Error(), "snippet is nil") {
		t.Errorf("expected snippet error, got %v", err)
	}

	// Verify clipboard was not written
	if mockClipboard.Content != "" {
		t.Errorf("expected clipboard to be empty, got %q", mockClipboard.Content)
	}
}

func TestRunCopy_SnippetNotFoundByTitle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	mockClipboard := &services.MockClipboard{}
	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &strings.Builder{},
		Stderr:    &strings.Builder{},
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{}, "NonExistent Title", 0)
	if err == nil {
		t.Fatal("expected error for non-existent snippet")
	}
	if !strings.Contains(err.Error(), "snippet not found") {
		t.Errorf("expected 'snippet not found' error, got %v", err)
	}
}

func TestRunCopy_ClipboardFailure(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, _ = setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	// Mock clipboard that returns an error
	mockClipboard := &services.MockClipboard{
		ErrorToReturn: errors.New("clipboard unavailable"),
	}

	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &strings.Builder{},
		Stderr:    &strings.Builder{},
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{"1"}, "", 0)
	if err == nil {
		t.Fatal("expected error from clipboard failure")
	}
	if !strings.Contains(err.Error(), "failed to copy to clipboard") {
		t.Errorf("expected clipboard error, got %v", err)
	}
}

func TestRunCopy_UsesWritersFromOpts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, _, _ = setupCopyTestDB(t, dbPath)
	defer database.CloseDB()

	mockClipboard := &services.MockClipboard{}
	var stdout, stderr strings.Builder

	opts := &CLIOptions{
		DBPath:    dbPath,
		Stdout:    &stdout,
		Stderr:    &stderr,
		Clipboard: mockClipboard,
	}

	err := runCopy(opts, []string{"1"}, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output went to custom stdout
	if stdout.String() == "" {
		t.Error("expected output in custom stdout")
	}

	// Verify nothing in stderr
	if stderr.String() != "" {
		t.Errorf("expected empty stderr, got %q", stderr.String())
	}
}

func TestParseID_Valid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"single digit", "1", 1, false},
		{"multiple digits", "12345", 12345, false},
		{"large number", "9223372036854775807", 9223372036854775807, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"zero", "0"},
		{"negative", "-1"},
		{"non-numeric", "abc"},
		{"float", "1.5"},
		{"with spaces", " 123 "},
		{"with letters", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseID(tt.input)
			if err == nil {
				t.Error("expected error for invalid input")
			}
		})
	}
}

func TestNewCopyCmd_HelpOutput(t *testing.T) {
	cmd := newCopyCmd(&CLIOptions{})

	// Verify command properties
	if cmd.Use != "copy [id]" {
		t.Errorf("expected Use 'copy [id]', got %q", cmd.Use)
	}
	if cmd.Short != "Copy a snippet to clipboard" {
		t.Errorf("expected Short 'Copy a snippet to clipboard', got %q", cmd.Short)
	}

	// Verify flags
	titleFlag := cmd.Flags().Lookup("title")
	if titleFlag == nil {
		t.Fatal("expected --title flag")
	}
	if titleFlag.Usage != "Copy snippet by exact title match" {
		t.Errorf("unexpected title flag usage: %q", titleFlag.Usage)
	}
}
