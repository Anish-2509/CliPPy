package services

import (
	"clippy/internal/database"
	"clippy/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Custom errors for snippet operations
var (
	// ErrSnippetNotFound is returned when a snippet cannot be found
	ErrSnippetNotFound = errors.New("snippet not found")
	// ErrInvalidSnippetID is returned when the ID is invalid
	ErrInvalidSnippetID = errors.New("invalid snippet ID")
)

// SnippetSaver defines the interface for saving snippets
// This allows the CLI to depend on an interface rather than concrete implementation
type SnippetSaver interface {
	Save(snippet *models.Snippet) (int64, error)
}

// SnippetService handles business logic for snippets
type SnippetService struct {
	// db connection is accessed via database.DB singleton
}

// NewSnippetService creates a new snippet service
func NewSnippetService() *SnippetService {
	return &SnippetService{}
}

// Save saves a snippet to the database and returns the new ID
func (s *SnippetService) Save(snippet *models.Snippet) (int64, error) {
	// Ensure database is initialized
	if database.DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	// Validate the snippet before saving
	if err := snippet.Validate(); err != nil {
		return 0, fmt.Errorf("validation failed: %w", err)
	}

	// Insert the snippet
	result, err := database.DB.Exec(
		`INSERT INTO snippets (title, code, language, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		snippet.Title,
		snippet.Code,
		snippet.Language,
		snippet.TagsString(),
		snippet.CreatedAt,
		snippet.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert snippet: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// List returns all snippets, optionally filtered by tags
// If tagFilters is non-empty, only snippets matching at least one of the tags are returned
func (s *SnippetService) List(tagFilters []string) ([]models.Snippet, error) {
	// Ensure database is initialized
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Query all snippets, ordered by created_at descending (newest first)
	query := `
		SELECT id, title, code, language, tags, created_at, updated_at
		FROM snippets
		ORDER BY created_at DESC, id DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query snippets: %w", err)
	}
	defer rows.Close()

	var snippets []models.Snippet
	for rows.Next() {
		var s models.Snippet
		var tagsStr string
		var createdAt, updatedAt interface{}

		// Try scanning timestamps as interface{} to handle multiple types
		err := rows.Scan(&s.ID, &s.Title, &s.Code, &s.Language, &tagsStr, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snippet: %w", err)
		}

		// Parse tags from comma-separated string
		s.Tags = models.ScanTags(tagsStr)

		// Parse timestamps - try direct time.Time scan first, then string fallback
		s.CreatedAt, err = scanTimestamp(createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at %v: %w", createdAt, err)
		}

		s.UpdatedAt, err = scanTimestamp(updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at %v: %w", updatedAt, err)
		}

		snippets = append(snippets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snippets: %w", err)
	}

	// Apply tag filters if provided
	if len(tagFilters) > 0 {
		snippets = filterByTags(snippets, tagFilters)
	}

	return snippets, nil
}

// filterByTags filters snippets to only those matching at least one of the provided tags
// Matching is case-insensitive
func filterByTags(snippets []models.Snippet, tagFilters []string) []models.Snippet {
	var filtered []models.Snippet

	for _, snippet := range snippets {
		// Check if snippet matches any of the filter tags
		if matchesAnyTag(snippet, tagFilters) {
			filtered = append(filtered, snippet)
		}
	}

	return filtered
}

// matchesAnyTag returns true if the snippet contains at least one of the filter tags
func matchesAnyTag(snippet models.Snippet, tagFilters []string) bool {
	for _, filter := range tagFilters {
		filter = strings.ToLower(strings.TrimSpace(filter))
		for _, tag := range snippet.Tags {
			if strings.ToLower(tag) == filter {
				return true
			}
		}
	}
	return false
}

// Search searches for snippets matching the query in title or tags, optionally filtered by language
// Query matching is case-insensitive partial match on title and tags
// Language filter is case-insensitive exact match
func (s *SnippetService) Search(query string, langFilter string) ([]models.Snippet, error) {
	// Ensure database is initialized
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate and normalize query
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// Normalize language filter
	langFilter = strings.TrimSpace(strings.ToLower(langFilter))

	// Get all snippets first (in-memory filtering for correctness)
	allSnippets, err := s.List(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve snippets: %w", err)
	}

	// Filter by search query and language
	var results []models.Snippet
	for _, snippet := range allSnippets {
		if matchesSearchQuery(snippet, query, langFilter) {
			results = append(results, snippet)
		}
	}

	return results, nil
}

// matchesSearchQuery returns true if the snippet matches the search criteria
func matchesSearchQuery(snippet models.Snippet, query string, langFilter string) bool {
	// Check language filter first (if specified)
	if langFilter != "" && snippet.Language != langFilter {
		return false
	}

	// Case-insensitive query matching
	queryLower := strings.ToLower(query)

	// Check if query matches title (partial match)
	if strings.Contains(strings.ToLower(snippet.Title), queryLower) {
		return true
	}

	// Check if query matches any tag (partial match)
	for _, tag := range snippet.Tags {
		if strings.Contains(strings.ToLower(tag), queryLower) {
			return true
		}
	}

	return false
}

// scanTimestamp converts an interface{} timestamp value to time.Time
// Handles both direct time.Time scans and string fallbacks
func scanTimestamp(ts interface{}) (time.Time, error) {
	// If it's already time.Time, return it directly
	if t, ok := ts.(time.Time); ok {
		return t, nil
	}

	// If it's a string, parse it
	tsStr, ok := ts.(string)
	if !ok {
		// For other types (like []byte), convert to string
		if b, ok := ts.([]byte); ok {
			tsStr = string(b)
		} else {
			return time.Time{}, fmt.Errorf("unsupported timestamp type: %T", ts)
		}
	}

	return parseTimestampString(tsStr)
}

// parseTimestampString parses a timestamp string from SQLite
// SQLite can return timestamps in various formats, so we try common ones
func parseTimestampString(ts string) (time.Time, error) {
	// Try RFC3339 format (with or without nanoseconds)
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t, nil
	}

	// Try SQLite's default format: "2006-01-02 15:04:05"
	if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
		return t, nil
	}

	// Try format with fractional seconds: "2006-01-02 15:04:05.999999999"
	if t, err := time.Parse("2006-01-02 15:04:05.999999999", ts); err == nil {
		return t, nil
	}

	// Try ISO 8601-like format: "2006-01-02T15:04:05"
	if t, err := time.Parse("2006-01-02T15:04:05", ts); err == nil {
		return t, nil
	}

	// Last resort: just try parsing with Go's default layout
	return time.Parse(time.RFC3339, ts)
}

// GetById retrieves a snippet by ID
func (s *SnippetService) GetById(id int64) (*models.Snippet, error) {
	// Ensure database is initialized
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate ID
	if id <= 0 {
		return nil, ErrInvalidSnippetID
	}

	query := `
		SELECT id, title, code, language, tags, created_at, updated_at
		FROM snippets
		WHERE id = ?
	`

	var snippet models.Snippet
	var tagsStr string
	var createdAt, updatedAt interface{}

	err := database.DB.QueryRow(query, id).Scan(&snippet.ID, &snippet.Title, &snippet.Code, &snippet.Language, &tagsStr, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSnippetNotFound
		}
		return nil, fmt.Errorf("failed to query snippet: %w", err)
	}

	// Parse tags
	snippet.Tags = models.ScanTags(tagsStr)

	// Parse timestamps
	snippet.CreatedAt, err = scanTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	snippet.UpdatedAt, err = scanTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &snippet, nil
}

// GetByTitle retrieves a snippet by exact title match
// Returns the first matching snippet if multiple exist
func (s *SnippetService) GetByTitle(title string) (*models.Snippet, error) {
	// Ensure database is initialized
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate title
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	query := `
		SELECT id, title, code, language, tags, created_at, updated_at
		FROM snippets
		WHERE title = ?
		LIMIT 1
	`

	var snippet models.Snippet
	var tagsStr string
	var createdAt, updatedAt interface{}

	err := database.DB.QueryRow(query, title).Scan(&snippet.ID, &snippet.Title, &snippet.Code, &snippet.Language, &tagsStr, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSnippetNotFound
		}
		return nil, fmt.Errorf("failed to query snippet: %w", err)
	}

	// Parse tags
	snippet.Tags = models.ScanTags(tagsStr)

	// Parse timestamps
	snippet.CreatedAt, err = scanTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	snippet.UpdatedAt, err = scanTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &snippet, nil
}

// Delete removes a snippet by ID
func (s *SnippetService) Delete(id int64) error {
	// TODO: Implement database delete
	return nil
}
