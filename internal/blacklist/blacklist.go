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

// IsBlacklisted checks if a command matches any blacklist pattern.
//
// Matching rules:
//  1. The pattern is tokenized into words. Each word in the pattern must match
//     the corresponding word in the command (in order from the beginning).
//  2. A trailing "*" in the pattern matches any remaining words (including none).
//     e.g. "rm *" matches "rm -rf /" and also "rm file.txt", but NOT "rm" alone.
//  3. Without a trailing "*", all pattern words must exactly match (the command
//     may have extra arguments after the matched words — trailing args are ignored
//     so "rm -rf" matches "rm -rf /" as well as "rm -rf").
//  4. The first word of both pattern and command is compared by base name, so
//     "/usr/bin/rm" is treated the same as "rm".
func (c *Checker) IsBlacklisted(command string) (bool, string) {
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return false, ""
	}

	// Normalize the command's first word to its base name (e.g. /usr/bin/rm -> rm)
	cmdParts[0] = filepath.Base(cmdParts[0])

	for _, pattern := range c.blacklist {
		if matchPattern(pattern, cmdParts) {
			return true, pattern
		}
	}

	return false, ""
}

// matchPattern reports whether cmdParts matches the given pattern.
func matchPattern(pattern string, cmdParts []string) bool {
	patParts := strings.Fields(pattern)
	if len(patParts) == 0 {
		return false
	}

	// Normalize pattern's first word to base name as well
	patParts[0] = filepath.Base(patParts[0])

	hasTrailingWildcard := patParts[len(patParts)-1] == "*"
	if hasTrailingWildcard {
		// Remove the trailing "*" from consideration
		patParts = patParts[:len(patParts)-1]
	}

	// With a trailing wildcard the command must have MORE words than the pattern
	// prefix (the "*" requires at least one additional argument).
	// Without a wildcard the command must have at least as many words as the pattern
	// (extra trailing args are also allowed, e.g. "rm -rf" matches "rm -rf /extra").
	if hasTrailingWildcard {
		if len(cmdParts) <= len(patParts) {
			return false
		}
	} else {
		if len(cmdParts) < len(patParts) {
			return false
		}
	}

	// Match each pattern word against the corresponding command word
	for i, p := range patParts {
		if cmdParts[i] != p {
			return false
		}
	}

	return true
}

// FormatBlacklistForPrompt formats the blacklist for inclusion in LLM prompts
func (c *Checker) FormatBlacklistForPrompt() string {
	if len(c.blacklist) == 0 {
		return ""
	}

	return "Command Blacklist:\n- " + strings.Join(c.blacklist, "\n- ")
}
