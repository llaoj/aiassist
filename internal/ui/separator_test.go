package ui

import (
	"strings"
	"testing"
)

func TestSeparator(t *testing.T) {
	sep := Separator()

	if sep == "" {
		t.Error("Expected separator to not be empty")
	}

	// Should contain dash characters
	if !strings.Contains(sep, "-") {
		t.Error("Expected separator to contain '-' characters")
	}

	// Check length is reasonable
	if len(sep) < 10 {
		t.Errorf("Expected separator to be at least 10 characters, got %d", len(sep))
	}
}
