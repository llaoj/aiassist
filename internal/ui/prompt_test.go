package ui

import (
	"path/filepath"
	"testing"
)

func TestHistory(t *testing.T) {
	// Create a temporary file for history
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Test creating new history
	h := NewHistory(historyFile, 100)
	if h.Size() != 0 {
		t.Errorf("Expected empty history, got %d entries", h.Size())
	}

	// Test appending entries
	entries := []string{"command1", "command2", "command3"}
	for _, entry := range entries {
		if err := h.Append(entry); err != nil {
			t.Errorf("Failed to append entry: %v", err)
		}
	}

	if h.Size() != 3 {
		t.Errorf("Expected 3 entries, got %d", h.Size())
	}

	// Test saving
	if err := h.Save(); err != nil {
		t.Errorf("Failed to save history: %v", err)
	}

	// Test loading into a new history instance
	h2 := NewHistory(historyFile, 100)
	if err := h2.Load(); err != nil {
		t.Errorf("Failed to load history: %v", err)
	}

	if h2.Size() != 3 {
		t.Errorf("Expected 3 entries after load, got %d", h2.Size())
	}

	// Verify entries match
	loaded := h2.GetEntries()
	for i, expected := range entries {
		if loaded[i] != expected {
			t.Errorf("Entry %d mismatch: expected %q, got %q", i, expected, loaded[i])
		}
	}
}

func TestHistoryMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create history with max size of 5
	h := NewHistory(historyFile, 5)

	// Add 10 entries
	for i := 0; i < 10; i++ {
		h.Append(string(rune('0' + i)))
	}

	// Should only keep the last 5
	if h.Size() != 5 {
		t.Errorf("Expected 5 entries (max size), got %d", h.Size())
	}

	// Verify it's the last 5 entries
	entries := h.GetEntries()
	for i, expected := range []string{"5", "6", "7", "8", "9"} {
		if entries[i] != expected {
			t.Errorf("Entry %d mismatch: expected %q, got %q", i, expected, entries[i])
		}
	}
}

func TestHistoryDuplicateConsecutive(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	h := NewHistory(historyFile, 100)

	// Add same entry twice
	h.Append("command")
	h.Append("command")

	// Should only have one entry
	if h.Size() != 1 {
		t.Errorf("Expected 1 entry (no duplicates), got %d", h.Size())
	}
}

func TestHistoryEmptyEntry(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	h := NewHistory(historyFile, 100)

	// Add empty entry
	h.Append("")
	h.Append("   ")

	// Should have no entries
	if h.Size() != 0 {
		t.Errorf("Expected 0 entries (empty strings), got %d", h.Size())
	}
}

func TestHistoryGetRecent(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	h := NewHistory(historyFile, 100)

	// Add 10 entries
	for i := 0; i < 10; i++ {
		h.Append(string(rune('0' + i)))
	}

	// Get recent 3
	recent := h.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent entries, got %d", len(recent))
	}

	// Should be the last 3
	expected := []string{"7", "8", "9"}
	for i, exp := range expected {
		if recent[i] != exp {
			t.Errorf("Recent entry %d mismatch: expected %q, got %q", i, exp, recent[i])
		}
	}
}

func TestIsTerminal(t *testing.T) {
	// This test just verifies the function doesn't panic
	// In a test environment, stdout is likely not a terminal
	result := IsTerminal()
	t.Logf("IsTerminal: %v", result)
}

func TestHistoryClear(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	h := NewHistory(historyFile, 100)

	// Add entries
	for i := 0; i < 5; i++ {
		h.Append(string(rune('0' + i)))
	}

	// Clear
	if err := h.Clear(); err != nil {
		t.Errorf("Failed to clear history: %v", err)
	}

	if h.Size() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", h.Size())
	}
}