package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrEmptyCode is returned when code content is empty
	ErrEmptyCode = errors.New("code content cannot be empty")
	// ErrInvalidTag is returned when a tag contains invalid characters
	ErrInvalidTag = errors.New("tag contains invalid characters (only alphanumeric, dash, underscore allowed)")
	// ErrTitleTooLong is returned when title exceeds max length
	ErrTitleTooLong = errors.New("title exceeds maximum length of 200 characters")
	// ErrCodeTooLong is returned when code exceeds max length
	ErrCodeTooLong = errors.New("code exceeds maximum length of 100000 characters")
)

// MaxLengths defines maximum field lengths
const (
	MaxTitleLength = 200
	MaxCodeLength  = 100000
	MaxTagLength   = 50
)

// Snippet represents a code snippet stored in the database
type Snippet struct {
	ID        int64
	Title     string
	Code      string
	Language  string
	Tags      []string // Stored as comma-separated in DB
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSnippet creates a new Snippet with validation
// If title is empty, generates a default title like "Untitled-2025-03-01-142530"
func NewSnippet(title, code, language string, tags []string) (*Snippet, error) {
	if err := validateCode(code); err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)

	// Generate default title if empty
	if title == "" {
		title = generateDefaultTitle()
	} else if err := validateTitle(title); err != nil {
		return nil, err
	}

	// Normalize and validate tags
	normalizedTags, err := normalizeTags(tags)
	if err != nil {
		return nil, err
	}

	// Normalize language (lowercase, trim)
	language = strings.TrimSpace(strings.ToLower(language))

	now := time.Now()
	return &Snippet{
		Title:     title,
		Code:      code,
		Language:  language,
		Tags:      normalizedTags,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate checks if the snippet is valid for saving/updating
// IMPORTANT: This method modifies the snippet in-place to apply normalization
func (s *Snippet) Validate() error {
	s.Title = strings.TrimSpace(s.Title)

	// Normalize and validate title
	if s.Title == "" {
		s.Title = generateDefaultTitle()
	} else if err := validateTitle(s.Title); err != nil {
		return err
	}

	// Validate code
	if err := validateCode(s.Code); err != nil {
		return err
	}

	// Normalize and validate tags in-place
	normalizedTags, err := normalizeTags(s.Tags)
	if err != nil {
		return err
	}
	s.Tags = normalizedTags

	// Normalize language in-place
	s.Language = strings.TrimSpace(strings.ToLower(s.Language))

	return nil
}

// TagsString returns tags as a comma-separated string for DB storage
func (s *Snippet) TagsString() string {
	if len(s.Tags) == 0 {
		return ""
	}
	return strings.Join(s.Tags, ",")
}

// SetTagsFromString parses comma-separated tags and sets them
func (s *Snippet) SetTagsFromString(tagsStr string) error {
	if tagsStr == "" {
		s.Tags = []string{}
		return nil
	}

	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	normalized, err := normalizeTags(tags)
	if err != nil {
		return err
	}
	s.Tags = normalized
	return nil
}

// HasTag returns true if the snippet has the given tag (case-insensitive)
func (s *Snippet) HasTag(tag string) bool {
	searchTag := strings.ToLower(tag)
	for _, t := range s.Tags {
		if strings.ToLower(t) == searchTag {
			return true
		}
	}
	return false
}

// MatchesTag returns true if any snippet tag matches the search term (case-insensitive partial match)
func (s *Snippet) MatchesTag(searchTerm string) bool {
	searchLower := strings.ToLower(searchTerm)
	for _, tag := range s.Tags {
		if strings.Contains(strings.ToLower(tag), searchLower) {
			return true
		}
	}
	return false
}

// TouchUpdatedAt updates the UpdatedAt timestamp to current time
func (s *Snippet) TouchUpdatedAt() {
	s.UpdatedAt = time.Now()
}

// validateTitle checks if title meets requirements
func validateTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if len(title) > MaxTitleLength {
		return fmt.Errorf("%w: length %d exceeds %d", ErrTitleTooLong, len(title), MaxTitleLength)
	}
	return nil
}

// validateCode checks if code meets requirements
func validateCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrEmptyCode
	}
	if len(code) > MaxCodeLength {
		return fmt.Errorf("%w: length %d exceeds %d", ErrCodeTooLong, len(code), MaxCodeLength)
	}
	return nil
}

// normalizeTags validates and normalizes tags
// Rules:
// - Convert to lowercase
// - Trim whitespace
// - Remove empty tags
// - Remove duplicates
// - Validate tag format (alphanumeric, dash, underscore only)
func normalizeTags(tags []string) ([]string, error) {
	if tags == nil {
		return []string{}, nil
	}

	// Tag validation regex: alphanumeric, dash, underscore allowed
	tagRegex := regexp.MustCompile(`^[a-z0-9_-]+$`)

	seen := make(map[string]bool)
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		// Trim and lowercase
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)

		// Skip empty
		if tag == "" {
			continue
		}

		// Check max length
		if len(tag) > MaxTagLength {
			return nil, fmt.Errorf("tag '%s' exceeds maximum length of %d", tag, MaxTagLength)
		}

		// Validate format
		if !tagRegex.MatchString(tag) {
			return nil, fmt.Errorf("%w: invalid tag '%s'", ErrInvalidTag, tag)
		}

		// Skip duplicates
		if !seen[tag] {
			seen[tag] = true
			normalized = append(normalized, tag)
		}
	}

	return normalized, nil
}

// generateDefaultTitle creates a default title using current timestamp
func generateDefaultTitle() string {
	return fmt.Sprintf("Untitled-%s", time.Now().Format("2006-01-02-150405"))
}

// ScanTags converts a DB string (comma-separated) to a tag slice
func ScanTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
