# CliPPy - Pending Tasks

## Completed Enhancements (Post-Verification)
- [x] Writer injection (Stdout/Stderr) for testable output
- [x] Timestamp hardening (direct time.Time scan with string fallback)
- [x] Consolidated parseTags tests into single canonical test suite
- [x] Helper functions stdout/stderr for safe writer access

## High Priority
- [x] Define Snippet model with validation in models/snippet.go
- [x] Setup Cobra CLI framework in main.go and commands.go
- [x] Implement Save command (--title, --lang, --tags, --code, stdin)
- [x] Handle missing title/tags - make title optional with auto-generated default
- [x] Implement List command with table output and tag filtering
- [x] Write database CRUD functions in snippet_service.go (Save and List implemented, Delete/GetById stubbed)

## Medium Priority
- [x] Implement Search command (title/tag search + lang filter)
- [x] Implement Copy command with clipboard integration
- [ ] Implement Edit command (open in editor + update timestamp)
- [ ] Implement Delete command (by ID/title + confirmation)
- [ ] Implement Editor detection and temp file handling
- [x] Implement Tag parsing and normalization utilities (already implemented in models)

## Feature Expansion
- [ ] Add favorites/pinning support (`save --favorite`, `list --favorite`)
- [ ] Add `recent` command for recently copied/edited snippets
- [ ] Add import/export commands (JSON + Markdown backup/restore)
- [ ] Add duplicate detection on save (title/content similarity warning)
- [ ] Add optional snippet metadata fields (`description`, `source`)
- [ ] Add soft-delete and restore workflow before permanent delete

## TUI Experience
- [ ] Introduce lightweight interactive mode using Bubble Tea
- [ ] Add `search --interactive` with live filtering while typing
- [ ] Add command palette for quick action switching
- [ ] Add result preview pane (full snippet/notes on selection)
- [ ] Add keyboard shortcuts (`j/k`, `enter`, `/`, `e`, `d`)
- [ ] Add status footer with current item context and shortcuts
- [ ] Add debounced realtime search updates for smooth UX
