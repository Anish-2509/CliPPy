# Snippy

A fast, offline command-line knowledge base for saving and reusing code snippets, commands, and notes. Snippy stores everything locally in SQLite and gives you instant search and clipboard copy so you can paste the answer and keep moving.

## Why Snippy

You have commands you look up over and over: `docker prune`, SQL joins, curl flags, git incantations. Snippy lets you save them once and retrieve them in seconds without opening a browser.

## Features (MVP)

- Save snippets with title, language, and tags
- List all snippets in a readable table
- Search by title or tag (case-insensitive)
- Copy a snippet to the clipboard by ID

## Planned Enhancements

- Fuzzy search for typos and partial matches
- Open snippet in `$EDITOR`
- Import/export snippets (JSON)
- Sync directory (opt-in) for teams

## Tech Stack

- Language: Go
- CLI framework: Cobra (recommended)
- Storage: SQLite
- Clipboard: platform-aware clipboard library

## Architecture

Snippy is organized into three layers for clarity and testability:

1. CLI layer
   - Parses commands and flags
   - Formats output and errors

2. Business logic layer
   - Validates inputs
   - Applies tagging and search rules
   - Coordinates clipboard copy and editor integration

3. Data layer
   - Manages SQLite connection
   - Handles schema migrations
   - Runs queries

## Database Schema

```sql
CREATE TABLE snippets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    code TEXT NOT NULL,
    language TEXT,
    tags TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## CLI Design

Examples of how the CLI will look:

- Save with an editor:
  - `snippy save --title "Docker Prune" --lang "bash" --tags "docker,cleanup"`
- Save directly with code:
  - `snippy save --title "JSON Parse" --lang "go" --tags "json" --code "fmt.Println(\"hi\")"`
- List all snippets:
  - `snippy list`
- Search snippets:
  - `snippy search "docker"`
- Copy snippet by ID:
  - `snippy copy 5`

## Save Command Flow

1. User runs `snippy save --title "JSON Parse" --lang go`
2. If `--code` is not provided, open `$EDITOR` with a temp file
3. Read the file contents after the editor closes
4. Insert into SQLite
5. Print success and the new ID

## Clipboard Behavior

- `snippy copy <id>` loads the snippet content and copies it to the system clipboard
- Should work on macOS, Windows, and Linux

## Go Implementation Plan

1. Initialize module
   - `go mod init snippy`
2. Set up CLI
   - `cobra init` and add subcommands: `save`, `list`, `search`, `copy`
3. Add SQLite layer
   - create connection
   - ensure schema exists
   - implement CRUD
4. Implement business logic
   - validation and tag normalization
   - search rules
5. Add clipboard support
   - use a cross-platform clipboard package
6. Add tests
   - unit tests for search and storage

## Fuzzy Search (Optional)

Use a fuzzy matcher so `dckr prun` still finds `Docker Prune`.

- Go library: `sahilm/fuzzy`

## Development Notes

- Keep data in a user-specific config dir (for example `~/.local/share/snippy/snippy.db` or `%AppData%\snippy\snippy.db`)
- Provide a `--db` flag for custom locations
- Keep output stable and scriptable (no color by default)

## License

Choose a license when publishing (MIT is a common default).
