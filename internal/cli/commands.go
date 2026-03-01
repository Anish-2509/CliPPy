package cli

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"clippy/internal/services"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CLIOptions holds configuration for CLI commands
type CLIOptions struct {
	// DBPath is the user-provided custom path via --db flag
	DBPath string
	// Stdin is injected for testing, defaults to os.Stdin
	Stdin io.Reader
	// Stdout is the output writer, defaults to os.Stdout
	Stdout io.Writer
	// Stderr is the error writer, defaults to os.Stderr
	Stderr io.Writer
	// Clipboard is the clipboard writer, defaults to RealClipboard
	Clipboard services.ClipboardWriter
}

// NewRootCmd creates and configures the root Cobra command
func NewRootCmd(opts *CLIOptions) *cobra.Command {
	if opts == nil {
		opts = &CLIOptions{}
	}

	rootCmd := &cobra.Command{
		Use:   "clippy",
		Short: "A CLI tool for clipboard management",
		Long: `CliPPy - Personal Offline Knowledge Base

A command-line tool for saving, organizing, and retrieving code snippets.
Store your commands, code patterns, and notes locally and retrieve them instantly.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&opts.DBPath, "db", "", "Custom path to clippy database")

	// Add subcommands
	rootCmd.AddCommand(
		newSaveCmd(opts),
		newListCmd(opts),
		newSearchCmd(opts),
		newCopyCmd(opts),
		newEditCmd(opts),
		newDeleteCmd(opts),
	)

	return rootCmd
}

// newSaveCmd creates the save command
func newSaveCmd(opts *CLIOptions) *cobra.Command {
	var title, lang, tags, code string

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save a new code snippet",
		Long: `Save a new code snippet with title, language, and tags.

Code input precedence:
  1. --code flag (highest priority)
  2. stdin (piped input only, not interactive terminal)
  3. Error (if neither provided)

Examples:
  clippy save --title "Docker Prune" --lang bash --tags "docker,cleanup" --code "docker system prune -af"
  echo "docker system prune -af" | clippy save --title "Docker Prune" --lang bash`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSave(opts, title, lang, tags, code)
		},
	}

	// Command flags
	cmd.Flags().StringVar(&title, "title", "", "Snippet title (auto-generated if empty)")
	cmd.Flags().StringVar(&lang, "lang", "", "Programming language")
	cmd.Flags().StringVar(&lang, "language", "", "Programming language (shorthand for --lang)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags (e.g., 'docker,cleanup,devops')")
	cmd.Flags().StringVar(&code, "code", "", "Code content (use stdin or omit for editor)")

	return cmd
}

// runSave executes the save command logic
func runSave(opts *CLIOptions, title, lang, tags, code string) error {
	// Resolve code source: --code flag takes precedence over stdin
	finalCode, err := resolveCodeSource(opts.Stdin, code)
	if err != nil {
		return err
	}

	// Parse tags
	tagSlice := parseTags(tags)

	// Create snippet with validation
	snippet, err := models.NewSnippet(title, finalCode, lang, tagSlice)
	if err != nil {
		return fmt.Errorf("invalid snippet: %w", err)
	}

	if err := initDB(opts); err != nil {
		return err
	}
	defer cleanupDB()

	// Create service and save
	snippetService := services.NewSnippetService()
	id, err := snippetService.Save(snippet)
	if err != nil {
		return fmt.Errorf("failed to save snippet: %w", err)
	}

	// Output success
	fmt.Fprintf(stdout(opts), "Saved snippet with ID: %d\n", id)
	return nil
}

func initDB(opts *CLIOptions) error {
	if opts == nil {
		return errors.New("cli options not initialized")
	}
	if err := database.InitDB(opts.DBPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

// cleanupDB closes the database connection
func cleanupDB() error {
	return database.CloseDB()
}

// stdout returns the stdout writer from opts, falling back to os.Stdout
func stdout(opts *CLIOptions) io.Writer {
	if opts != nil && opts.Stdout != nil {
		return opts.Stdout
	}
	return os.Stdout
}

// stderr returns the stderr writer from opts, falling back to os.Stderr
func stderr(opts *CLIOptions) io.Writer {
	if opts != nil && opts.Stderr != nil {
		return opts.Stderr
	}
	return os.Stderr
}

// runList executes the list command logic
func runList(opts *CLIOptions, tagsFilter string) error {
	if err := initDB(opts); err != nil {
		return err
	}
	defer cleanupDB()

	// Parse tag filters
	tagFilters := parseTags(tagsFilter)

	// Get snippets from service
	snippetService := services.NewSnippetService()
	snippets, err := snippetService.List(tagFilters)
	if err != nil {
		return fmt.Errorf("failed to list snippets: %w", err)
	}

	// Render snippets as table
	renderSnippetsTable(stdout(opts), snippets)

	return nil
}

// renderSnippetsTable renders snippets as a formatted table
func renderSnippetsTable(w io.Writer, snippets []models.Snippet) {
	if w == nil {
		w = os.Stdout
	}

	if len(snippets) == 0 {
		fmt.Fprintln(w, "No snippets found.")
		return
	}

	// Print header
	fmt.Fprintf(w, "%-6s %-30s %-15s %-20s %s\n", "ID", "Title", "Language", "Tags", "Created")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	// Print rows
	for _, s := range snippets {
		tagsStr := strings.Join(s.Tags, ",")
		if tagsStr == "" {
			tagsStr = "-"
		}
		// Truncate title if too long
		title := s.Title
		if len(title) > 27 {
			title = title[:27] + "..."
		}
		// Truncate tags if too long
		if len(tagsStr) > 17 {
			tagsStr = tagsStr[:17] + "..."
		}
		createdStr := s.CreatedAt.Format("2006-01-02")
		fmt.Fprintf(w, "%-6d %-30s %-15s %-20s %s\n", s.ID, title, s.Language, tagsStr, createdStr)
	}
}

// resolveCodeSource determines the code content based on precedence:
// 1. --code flag (if non-empty)
// 2. stdin (if piped content available, NOT interactive terminal)
// 3. Error (no code source available)
func resolveCodeSource(stdin io.Reader, codeFlag string) (string, error) {
	// If --code flag is provided and non-empty, use it
	if strings.TrimSpace(codeFlag) != "" {
		return codeFlag, nil
	}

	// Check if stdin is a piped input (not interactive terminal)
	if !isInputPiped(stdin) {
		return "", errors.New("code content required: provide --code flag or pipe via stdin")
	}

	// Read from stdin
	stdinContent, err := readStdin(stdin)
	if err != nil {
		return "", err
	}

	if stdinContent != "" {
		return stdinContent, nil
	}

	// No code source available
	return "", errors.New("code content required: provide --code flag or pipe via stdin")
}

// isInputPiped returns true if stdin is receiving piped input (not an interactive terminal)
func isInputPiped(stdin io.Reader) bool {
	if stdin == nil {
		return false
	}

	// Check if stdin is an *os.File (which it should be on real execution)
	f, ok := stdin.(*os.File)
	if !ok {
		// For testing with strings.Reader, assume it's piped
		return true
	}

	// Get file info to check if it's a terminal/character device
	info, err := f.Stat()
	if err != nil {
		return false
	}

	// Any non-terminal stdin (pipe or redirected file) is valid input source.
	return info.Mode()&os.ModeCharDevice == 0
}

// readStdin reads content from stdin
func readStdin(stdin io.Reader) (string, error) {
	if stdin == nil {
		return "", nil
	}

	stdinBytes, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	content := string(stdinBytes)
	// Trim whitespace for checking
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	return content, nil
}

// parseTags converts comma-separated tag string to slice
func parseTags(tags string) []string {
	if tags == "" {
		return []string{}
	}

	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// newListCmd creates the list command
func newListCmd(opts *CLIOptions) *cobra.Command {
	var tags string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved snippets",
		Long: `List all saved snippets in a formatted table. Optionally filter by tags.

Examples:
  clippy list
  clippy list --tags docker
  clippy list --tags docker,cleanup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts, tags)
		},
	}

	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags to filter snippets")

	return cmd
}

// newSearchCmd creates the search command
func newSearchCmd(opts *CLIOptions) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search snippets by title or tags",
		Long: `Search for snippets matching the query in title or tags. Optionally filter by language.

The search is case-insensitive and matches partial text in titles and tags.

Examples:
  clippy search docker
  clippy search prune --lang bash
  clippy search http --language go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(opts, args[0], lang)
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "", "Filter results by language")
	cmd.Flags().StringVar(&lang, "language", "", "Filter results by language (shorthand for --lang)")

	return cmd
}

// runSearch executes the search command logic
func runSearch(opts *CLIOptions, query, langFilter string) error {
	if strings.TrimSpace(query) == "" {
		return errors.New("search query cannot be empty")
	}

	if err := initDB(opts); err != nil {
		return err
	}
	defer cleanupDB()

	// Get snippets from service
	snippetService := services.NewSnippetService()
	snippets, err := snippetService.Search(query, langFilter)
	if err != nil {
		return fmt.Errorf("failed to search snippets: %w", err)
	}

	// Check for empty results
	if len(snippets) == 0 {
		fmt.Fprintln(stdout(opts), "No matching snippets found.")
		return nil
	}

	// Render snippets using the same table renderer as list
	renderSnippetsTable(stdout(opts), snippets)

	return nil
}

// newCopyCmd creates the copy command
func newCopyCmd(opts *CLIOptions) *cobra.Command {
	var title string
	var id int64

	cmd := &cobra.Command{
		Use:   "copy [id]",
		Short: "Copy a snippet to clipboard",
		Long: `Copy a snippet's code to the system clipboard by ID or title.

Examples:
  clippy copy 5
  clippy copy --title "Docker Prune"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopy(opts, args, title, id)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Copy snippet by exact title match")
	cmd.Flags().Int64Var(&id, "id", 0, "Snippet ID (alternative to --title)")

	return cmd
}

// runCopy executes the copy command logic
func runCopy(opts *CLIOptions, args []string, title string, id int64) error {
	// Validate: either positional ID or --title flag, but not both
	hasID := len(args) > 0
	hasTitle := title != ""

	if hasID && hasTitle {
		return errors.New("cannot use both [id] and --title together")
	}
	if !hasID && !hasTitle {
		return errors.New("either [id] or --title is required")
	}

	if err := initDB(opts); err != nil {
		return err
	}
	defer cleanupDB()

	snippetService := services.NewSnippetService()
	var snippet *models.Snippet
	var err error

	// Resolve snippet by ID or title
	if hasID {
		// Parse ID from positional argument
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		snippet, err = snippetService.GetById(id)
	} else {
		snippet, err = snippetService.GetByTitle(title)
	}

	if err != nil {
		if errors.Is(err, services.ErrSnippetNotFound) {
			return fmt.Errorf("snippet not found")
		}
		return fmt.Errorf("failed to retrieve snippet: %w", err)
	}

	// Defensive check - snippet should never be nil if err is nil
	if snippet == nil {
		return fmt.Errorf("snippet is nil after retrieval")
	}

	// Ensure clipboard is initialized
	if opts.Clipboard == nil {
		return fmt.Errorf("clipboard not initialized")
	}

	// Write code to clipboard
	if err := opts.Clipboard.WriteAll(snippet.Code); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	// Output success
	fmt.Fprintf(stdout(opts), "Copied snippet '%s' (ID: %d) to clipboard.\n", snippet.Title, snippet.ID)
	return nil
}

// parseID parses a string ID into int64
func parseID(idStr string) (int64, error) {
	var id int64
	// Reject input with leading/trailing whitespace
	if strings.TrimSpace(idStr) != idStr {
		return 0, fmt.Errorf("invalid ID '%s': must be a positive integer", idStr)
	}
	// First, try to parse as integer
	n, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("invalid ID '%s': must be a positive integer", idStr)
	}
	if id <= 0 {
		return 0, fmt.Errorf("invalid ID '%s': must be a positive integer", idStr)
	}
	// Verify the entire string was consumed (no extra characters like "1.5" or "1abc")
	// Format the parsed ID back and compare with original input
	formatted := fmt.Sprintf("%d", id)
	if idStr != formatted {
		return 0, fmt.Errorf("invalid ID '%s': must be a positive integer", idStr)
	}
	return id, nil
}

// newEditCmd creates the edit command stub
func newEditCmd(opts *CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit an existing snippet",
		Long:  `Open a snippet in your default editor for modification. Updates the timestamp.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("edit command not yet implemented")
			}
			return fmt.Errorf("edit command not yet implemented (id: %s)", args[0])
		},
	}

	return cmd
}

// newDeleteCmd creates the delete command stub
func newDeleteCmd(opts *CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"del"},
		Short:   "Delete a snippet",
		Long:    `Delete a snippet by ID or title. Requires confirmation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("delete command not yet implemented")
			}
			return fmt.Errorf("delete command not yet implemented (id: %s)", args[0])
		},
	}

	return cmd
}

// Execute runs the root command with proper error handling
func Execute(opts *CLIOptions) {
	if opts == nil {
		opts = &CLIOptions{}
	}

	// Set default writers if not provided
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	// Set default clipboard if not provided
	if opts.Clipboard == nil {
		opts.Clipboard = &services.RealClipboard{}
	}

	rootCmd := NewRootCmd(opts)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(opts.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
