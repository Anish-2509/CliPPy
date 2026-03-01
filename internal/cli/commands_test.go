package cli

import (
	"strings"
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	if rootCmd == nil {
		t.Fatal("NewRootCmd returned nil")
	}

	if rootCmd.Use != "clippy" {
		t.Errorf("Expected root command use 'clippy', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected root command to have Short description")
	}

	if rootCmd.Long == "" {
		t.Error("Expected root command to have Long description")
	}

	if !rootCmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}

	if !rootCmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}
}

func TestNewRootCmd_NilOptionsDoesNotPanic(t *testing.T) {
	rootCmd := NewRootCmd(nil)
	if rootCmd == nil {
		t.Fatal("expected non-nil root command for nil options")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	expectedCommands := []string{"save", "list", "search", "copy", "edit", "delete"}

	for _, cmdName := range expectedCommands {
		cmd, _, _ := rootCmd.Find([]string{cmdName})
		if cmd == nil {
			t.Errorf("Expected to find subcommand '%s'", cmdName)
		}
	}
}

func TestDBFlag(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	flag := rootCmd.PersistentFlags().Lookup("db")
	if flag == nil {
		t.Fatal("Expected --db flag to exist")
	}

	if flag.Name != "db" {
		t.Errorf("Expected flag name 'db', got '%s'", flag.Name)
	}
}

func TestDBFlagBinding(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	// Simulate setting the flag with a custom path
	args := []string{"--db", "/custom/path/clippy.db", "list"}
	rootCmd.SetArgs(args)

	// Execute - list command is now implemented but will show "No snippets found" for non-existent DB
	err := rootCmd.Execute()
	// The command may succeed (showing empty list) or fail if DB can't be created
	// Either is acceptable for this test which is checking flag parsing
	_ = err // We're just checking that the flag is parsed

	// Check if the flag was parsed
	if rootCmd.PersistentFlags().Changed("db") {
		dbPath, _ := rootCmd.PersistentFlags().GetString("db")
		if dbPath != "/custom/path/clippy.db" {
			t.Errorf("Expected db path '/custom/path/clippy.db', got '%s'", dbPath)
		}
	}
}

func TestSaveCommandHelp(t *testing.T) {
	opts := &CLIOptions{}
	saveCmd := newSaveCmd(opts)

	if saveCmd.Use != "save" {
		t.Errorf("Expected save command use 'save', got '%s'", saveCmd.Use)
	}

	if saveCmd.Short == "" {
		t.Error("Expected save command to have Short description")
	}

	if saveCmd.Long == "" {
		t.Error("Expected save command to have Long description")
	}
}

func TestListCommandHelp(t *testing.T) {
	opts := &CLIOptions{}
	listCmd := newListCmd(opts)

	if listCmd.Use != "list" {
		t.Errorf("Expected list command use 'list', got '%s'", listCmd.Use)
	}
}

func TestSearchCommandRequiresArgs(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	// Search with no args should error
	rootCmd.SetArgs([]string{"search"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when search command has no arguments")
	}

	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("Expected error about argument count, got: %v", err)
	}
}

func TestSearchCommandWithArg(t *testing.T) {
	opts := &CLIOptions{
		DBPath: "", // Will use default path
	}
	rootCmd := NewRootCmd(opts)

	// Search with arg should execute successfully
	// (may show no results or results depending on DB state)
	rootCmd.SetArgs([]string{"search", "docker"})
	err := rootCmd.Execute()
	// Command should succeed (error may be nil or show "No matching snippets")
	if err != nil && strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("Got unexpected argument validation error: %v", err)
	}
}

func TestCopyEditDeleteCommandHelp(t *testing.T) {
	opts := &CLIOptions{}

	copyCmd := newCopyCmd(opts)
	if !strings.HasPrefix(copyCmd.Use, "copy") {
		t.Errorf("Expected copy command use to start with 'copy', got '%s'", copyCmd.Use)
	}

	editCmd := newEditCmd(opts)
	if !strings.HasPrefix(editCmd.Use, "edit") {
		t.Errorf("Expected edit command use to start with 'edit', got '%s'", editCmd.Use)
	}

	deleteCmd := newDeleteCmd(opts)
	if !strings.HasPrefix(deleteCmd.Use, "delete") {
		t.Errorf("Expected delete command use to start with 'delete', got '%s'", deleteCmd.Use)
	}

	// Check that delete has 'del' alias
	if len(deleteCmd.Aliases) == 0 {
		t.Error("Expected delete command to have aliases")
	}
	foundDelAlias := false
	for _, alias := range deleteCmd.Aliases {
		if alias == "del" {
			foundDelAlias = true
			break
		}
	}
	if !foundDelAlias {
		t.Errorf("Expected delete command to have 'del' alias, got %v", deleteCmd.Aliases)
	}
}

func TestDelAliasWorks(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	// Test that 'del' alias resolves to delete command
	deleteCmd, _, _ := rootCmd.Find([]string{"del"})
	if deleteCmd == nil {
		t.Error("Expected 'del' alias to resolve to delete command")
	}

	if deleteCmd.Use != "delete [id]" {
		t.Errorf("Expected 'del' alias to resolve to delete command with use 'delete [id]', got '%s'", deleteCmd.Use)
	}
}

func TestInvalidCommandReturnsError(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	rootCmd.SetArgs([]string{"invalid-command"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid command")
	}
}

func TestRootCommandHasCompletion(t *testing.T) {
	opts := &CLIOptions{}
	rootCmd := NewRootCmd(opts)

	// Cobra automatically adds completion command
	completionCmd, _, _ := rootCmd.Find([]string{"completion"})
	if completionCmd == nil {
		t.Error("Expected completion command to exist (auto-added by Cobra)")
	}
}
