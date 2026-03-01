# CliPPy - Personal Offline Knowledge Base

A TUI command-line tool built in Go that acts as a personal knowledge base for code snippets, commands, and notes. Users can save snippets with titles and tags, then retrieve them instantly without needing to search online.

## Why CliPPy

You have commands you look up over and over: `docker prune`, SQL joins, curl flags, git incantations. CliPPy lets you save them once and retrieve them in seconds without opening a browser.

## Architecture

The project follows a clean, layered architecture:

```
┌─────────────────┐
│   CLI Layer     │  (Cobra commands: save, list, search, copy, edit, delete)
└────────┬────────┘
         │
┌────────▼────────┐
│ Business Logic  │  (Validation, tagging, search algorithms)
└────────┬────────┘
         │
┌────────▼────────┐
│   Data Layer    │  (SQLite database operations)
└─────────────────┘
```

## Project Structure

```
clippy/
├── cmd/
│   └── clippy/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/
│   │   └── commands.go          # Cobra command definitions
│   ├── database/
│   │   ├── db.go                # Database connection & initialization
│   │   └── schema.go             # SQLite schema & migrations
│   ├── models/
│   │   └── snippet.go            # Snippet data model
│   └── services/
│       ├── snippet_service.go   # Business logic for snippets
│       └── clipboard.go         # Clipboard operations
├── go.mod
├── go.sum
└── README.md
```

## Database Schema

SQLite table structure:

```sql
CREATE TABLE snippets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    code TEXT NOT NULL,
    language TEXT,
    tags TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_title ON snippets(title);
CREATE INDEX IF NOT EXISTS idx_tags ON snippets(tags);
```

## Core Features

### 1. Save Command

- `clippy save --title "Docker Prune" --lang bash --tags "docker,cleanup"`
- If `--code` flag not provided, open temporary file in user's editor
- Support reading from stdin: `echo "code" | clippy save --title "..."`
- Save directly with code: `clippy save --title "JSON Parse" --lang "go" --tags "json" --code "fmt.Println(\"hi\")"`

### 2. List Command

- `clippy list` - Show all snippets in a formatted table
- `clippy list --tags "docker"` - Filter by tags
- Display: ID, Title, Language, Tags, Created Date

### 3. Search Command

- `clippy search "docker"` - Search in titles and tags
- `clippy search "prune" --lang bash` - Filter by language
- Optional: Fuzzy search using `sahilm/fuzzy` library

### 4. Copy Command

- `clippy copy 5` - Copy snippet #5 to clipboard
- `clippy copy --title "Docker Prune"` - Copy by title match

### Notes and Plain Text Support

- CliPPy already supports notes and any plain text, not only code.
- Internally, `code` stores arbitrary text content.
- This means `save`, `search`, `list`, and `copy` work for notes as well.
- Example:
  - `clippy save --title "Oncall Notes" --tags "notes,ops" --code "Restart order: API -> Worker -> Scheduler"`
  - `clippy copy --title "Oncall Notes"`

### 5. Edit Command

- `clippy edit 5` - Open snippet #5 in editor for modification
- Update `updated_at` timestamp

### 6. Delete Command

- `clippy delete 5` - Remove snippet by ID
- `clippy delete --title "..."` - Remove by title match
- Confirm before deletion

## Technical Stack

- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra)
- **Database:** [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Clipboard:** [atotto/clipboard](https://github.com/atotto/clipboard) or [golang.design/x/clipboard](https://golang.design/x/clipboard)
- **Table Formatting:** [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
- **Fuzzy Search (optional):** [sahilm/fuzzy](https://github.com/sahilm/fuzzy)
- **Editor Detection:** Use `$EDITOR` environment variable, fallback to `vim`/`nano`

## Implementation Details

### Data Storage

- Database file location: `~/.clippy/clippy.db` (or `%APPDATA%\clippy\clippy.db` on Windows)
- Create directory if it doesn't exist
- Initialize schema on first run

### Editor Integration

- Detect `$EDITOR` environment variable
- Fallback to platform-specific defaults (vim on Unix, notepad on Windows)
- Create temporary file, spawn editor process, read content after editor closes

### Tag Parsing

- Parse comma-separated tags: `"docker,cleanup,devops"`
- Normalize tags (lowercase, trim whitespace)
- Support tag filtering in list/search commands

### Clipboard Operations

- Cross-platform clipboard support
- Handle errors gracefully (e.g., if clipboard not available)

## Save Command Flow

1. User runs `clippy save --title "JSON Parse" --lang go`
2. If `--code` is not provided, open `$EDITOR` with a temp file
3. Read the file contents after the editor closes
4. Insert into SQLite
5. Print success and the new ID

## Development Notes

- Keep data in a user-specific config dir (`~/.clippy/clippy.db` or `%APPDATA%\clippy\clippy.db`)
- Provide a `--db` flag for custom locations
- Keep output stable and scriptable (no color by default)

## Product Ideas

### Practical Feature Upgrades

- Favorites / pinning:
  - Mark important snippets with `--favorite`, list with `clippy list --favorite`.
- Recent activity:
  - `clippy recent` to show recently copied/edited snippets.
- Import and export:
  - JSON/Markdown export and import for backup and sync.
- Duplicate detection:
  - Warn when saving nearly identical content/title.
- Rich metadata:
  - Add optional `description` and `source` fields.
- Safer deletion:
  - Soft delete + restore command before permanent purge.

### Realtime TUI Vibe (Lightweight, Legitimate)

- Interactive search mode:
  - `clippy search --interactive` with live filtering as user types.
- Live command palette:
  - A `:` style quick launcher for `save/list/search/copy/edit/delete`.
- Inline preview panel:
  - While navigating results, show full snippet/notes preview in side pane.
- Keyboard-first flow:
  - `j/k` navigation, `enter` copy, `/` filter, `e` edit, `d` delete.
- Live status hints:
  - Real-time footer showing selected item, tags, language, and shortcuts.
- Debounced search:
  - Update results after short typing delay for smooth UX.

### Suggested TUI Stack

- Bubble Tea (`charmbracelet/bubbletea`) for event loop
- Bubbles (`charmbracelet/bubbles`) for list/input/table components
- Lip Gloss (`charmbracelet/lipgloss`) for styling

## License

Choose a license when publishing (MIT is a common default).
