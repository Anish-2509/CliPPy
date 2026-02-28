# CliPPy - Pending Tasks

## High Priority
- [ ] Define Snippet model with validation in models/snippet.go
- [ ] Setup Cobra CLI framework in main.go and commands.go
- [ ] Implement Save command (--title, --lang, --tags, --code, stdin)
- [ ] Handle missing title/tags - make title optional with auto-generated default
- [ ] Implement List command with table output and tag filtering
- [ ] Write database CRUD functions in snippet_service.go

## Medium Priority
- [ ] Implement Search command (title/tag search + lang filter)
- [ ] Implement Copy command with clipboard integration
- [ ] Implement Edit command (open in editor + update timestamp)
- [ ] Implement Delete command (by ID/title + confirmation)
- [ ] Implement Editor detection and temp file handling
- [ ] Implement Tag parsing and normalization utilities
