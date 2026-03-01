package blacklist

import (
	"path/filepath"
	"strings"

	"github.com/llaoj/aiassist/internal/config"
)

// Checker checks commands against a blacklist
type Checker struct {
	blacklist []string
}

// NewChecker creates a new blacklist checker
func NewChecker() *Checker {
	cfg := config.Get()
	return &Checker{
		blacklist: cfg.GetBlacklist(),
	}
}

// IsBlacklisted checks if a command matches any blacklist pattern
// Supports glob patterns (e.g., "rm *", "kubectl delete *", "dd *")
func (c *Checker) IsBlacklisted(command string) (bool, string) {
	// Extract the base command (first word) for simple matching
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false, ""
	}

	baseCmd := parts[0]

	// Check each blacklist pattern
	for _, pattern := range c.blacklist {
		// If pattern ends with *, it should match any command starting with the prefix
		if strings.HasSuffix(pattern, "*") {
			// Get the prefix before *
			prefix := strings.TrimSuffix(pattern, "*")
			// Check if command starts with the prefix
			if strings.HasPrefix(command, prefix) {
				return true, pattern
			}
			// Also check base command
			if strings.HasPrefix(baseCmd, prefix) {
				return true, pattern
			}
		} else {
			// Exact match for patterns without wildcard
			if command == pattern || baseCmd == pattern {
				return true, pattern
			}
		}

		// Try traditional glob matching as well
		matched, err := filepath.Match(pattern, baseCmd)
		if err == nil && matched {
			return true, pattern
		}
	}

	return false, ""
}

// FormatBlacklistForPrompt formats the blacklist for inclusion in LLM prompts
func (c *Checker) FormatBlacklistForPrompt() string {
	if len(c.blacklist) == 0 {
		return ""
	}

	return "Command Blacklist:\n- " + strings.Join(c.blacklist, "\n- ")
}
