package util

import (
	"testing"
)

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		maxLen   int
		expected string
	}{
		{"short path unchanged", "/a/b", 10, "/a/b"},
		{"exact length unchanged", "12345", 5, "12345"},
		{"long path truncated", "/very/long/path/to/file.txt", 15, ".../to/file.txt"},
		{"minimal truncation", "abcdef", 5, "...ef"},
		{"single char over", "abcdef", 5, "...ef"},
		{"empty path", "", 5, ""},
		{"maxLen equals path length", "hello", 5, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncatePath(tt.path, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncatePath(%q, %d) = %q, want %q", tt.path, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"one byte", 1, "1 B"},
		{"exact 1 KB", 1024, "1.0 KB"},
		{"kilobytes", 1536, "1.5 KB"},
		{"exact 1 MB", 1024 * 1024, "1.0 MB"},
		{"megabytes", 2 * 1024 * 1024, "2.0 MB"},
		{"exact 1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"gigabytes", 3 * 1024 * 1024 * 1024, "3.0 GB"},
		{"just under 1 KB", 1023, "1023 B"},
		{"just under 1 MB", 1024*1024 - 1, "1024.0 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.size)
			if result != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}
