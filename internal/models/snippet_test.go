package models

import "testing"

func TestNewSnippet_TrimmedEmptyTitleGeneratesDefault(t *testing.T) {
	snippet, err := NewSnippet("   ", "echo hi", "BASH", []string{"Ops"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snippet.Title == "" {
		t.Fatal("expected generated title")
	}
	if snippet.Language != "bash" {
		t.Fatalf("expected language bash, got %q", snippet.Language)
	}
}

func TestValidate_NormalizesInPlace(t *testing.T) {
	snippet := &Snippet{
		Title:    "  My Title  ",
		Code:     "echo hi",
		Language: " Go ",
		Tags:     []string{" DevOps ", "devops", "CLI"},
	}

	if err := snippet.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snippet.Title != "My Title" {
		t.Fatalf("expected trimmed title, got %q", snippet.Title)
	}
	if snippet.Language != "go" {
		t.Fatalf("expected normalized language go, got %q", snippet.Language)
	}
	if len(snippet.Tags) != 2 {
		t.Fatalf("expected deduplicated tags length 2, got %d", len(snippet.Tags))
	}
	if snippet.Tags[0] != "devops" || snippet.Tags[1] != "cli" {
		t.Fatalf("unexpected tags: %#v", snippet.Tags)
	}
}
