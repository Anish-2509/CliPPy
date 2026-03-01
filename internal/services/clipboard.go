package services

import (
	"github.com/atotto/clipboard"
)

// ClipboardWriter defines the interface for writing to clipboard
// This allows for testable clipboard operations
type ClipboardWriter interface {
	WriteAll(text string) error
}

// RealClipboard is the production implementation using system clipboard
type RealClipboard struct{}

// WriteAll writes text to the system clipboard
func (rc *RealClipboard) WriteAll(text string) error {
	return clipboard.WriteAll(text)
}

// MockClipboard is a test double for clipboard operations
type MockClipboard struct {
	Content string
	// ErrorToReturn simulates a clipboard error
	ErrorToReturn error
}

// WriteAll stores content in the mock for testing
func (mc *MockClipboard) WriteAll(text string) error {
	if mc.ErrorToReturn != nil {
		return mc.ErrorToReturn
	}
	mc.Content = text
	return nil
}

// GetContent returns the mock clipboard content
func (mc *MockClipboard) GetContent() string {
	return mc.Content
}
