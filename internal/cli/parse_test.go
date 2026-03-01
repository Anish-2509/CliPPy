package cli

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		expected []string
	}{
		{
			name:     "empty string",
			tags:     "",
			expected: []string{},
		},
		{
			name:     "single tag",
			tags:     "docker",
			expected: []string{"docker"},
		},
		{
			name:     "multiple tags with spaces",
			tags:     "docker, cleanup, devops",
			expected: []string{"docker", "cleanup", "devops"},
		},
		{
			name:     "tags with empty entries",
			tags:     "docker,,cleanup",
			expected: []string{"docker", "cleanup"},
		},
		{
			name:     "whitespace only tags",
			tags:     "  ,  ,  ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.tags)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTags() = %v (len=%d), want %v (len=%d)", result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseTags()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
