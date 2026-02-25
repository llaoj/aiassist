package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// History manages command history with persistence
type History struct {
	filePath string
	maxSize  int
	mu       sync.RWMutex
	entries  []string
}

// NewHistory creates a new History instance
func NewHistory(filePath string, maxSize int) *History {
	return &History{
		filePath: filePath,
		maxSize:  maxSize,
		entries:  make([]string, 0),
	}
}

// Load loads history entries from file
func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start with empty history
			return nil
		}
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	h.entries = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Trim if exceeds max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}

	return nil
}

// Save saves history entries to file
func (h *History) Save() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range h.entries {
		if _, err := writer.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("failed to write history entry: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush history file: %w", err)
	}

	return nil
}

// Append adds a new entry to history
func (h *History) Append(entry string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry = strings.TrimSpace(entry)
	if entry == "" {
		return nil
	}

	// Don't add duplicate consecutive entries
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		return nil
	}

	h.entries = append(h.entries, entry)

	// Trim if exceeds max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:]
	}

	return nil
}

// GetEntries returns a copy of history entries
func (h *History) GetEntries() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := make([]string, len(h.entries))
	copy(entries, h.entries)
	return entries
}

// GetRecent returns the most recent n entries
func (h *History) GetRecent(n int) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.entries) {
		n = len(h.entries)
	}

	start := len(h.entries) - n
	if start < 0 {
		start = 0
	}

	entries := make([]string, n)
	copy(entries, h.entries[start:])
	return entries
}

// Clear clears all history entries
func (h *History) Clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]string, 0)
	return nil
}

// Size returns the number of history entries
func (h *History) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.entries)
}
