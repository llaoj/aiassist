package blacklist

import (
	"testing"
)

func TestIsBlacklisted(t *testing.T) {
	tests := []struct {
		name      string
		blacklist []string
		command   string
		want      bool
		pattern   string
	}{
		{
			name:      "Match rm command",
			blacklist: []string{"rm *"},
			command:   "rm -rf /",
			want:      true,
			pattern:   "rm *",
		},
		{
			name:      "Match kubectl delete",
			blacklist: []string{"kubectl delete *"},
			command:   "kubectl delete pod nginx",
			want:      true,
			pattern:   "kubectl delete *",
		},
		{
			name:      "Exact match shutdown",
			blacklist: []string{"shutdown"},
			command:   "shutdown",
			want:      true,
			pattern:   "shutdown",
		},
		{
			name:      "No match safe command",
			blacklist: []string{"rm *"},
			command:   "ls -la",
			want:      false,
			pattern:   "",
		},
		{
			name:      "Match dd command",
			blacklist: []string{"dd *"},
			command:   "dd if=/dev/zero of=/dev/sda",
			want:      true,
			pattern:   "dd *",
		},
		{
			name:      "Empty blacklist",
			blacklist: []string{},
			command:   "rm -rf /",
			want:      false,
			pattern:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &Checker{blacklist: tt.blacklist}
			got, pattern := checker.IsBlacklisted(tt.command)
			if got != tt.want {
				t.Errorf("IsBlacklisted() = %v, want %v", got, tt.want)
			}
			if got && pattern != tt.pattern {
				t.Errorf("IsBlacklisted() pattern = %v, want %v", pattern, tt.pattern)
			}
		})
	}
}

func TestFormatBlacklistForPrompt(t *testing.T) {
	tests := []struct {
		name      string
		blacklist []string
		want      string
	}{
		{
			name:      "Empty blacklist",
			blacklist: []string{},
			want:      "",
		},
		{
			name:      "Single item",
			blacklist: []string{"rm *"},
			want:      "Command Blacklist:\n- rm *",
		},
		{
			name:      "Multiple items",
			blacklist: []string{"rm *", "dd *", "kubectl delete *"},
			want:      "Command Blacklist:\n- rm *\n- dd *\n- kubectl delete *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &Checker{blacklist: tt.blacklist}
			got := checker.FormatBlacklistForPrompt()
			if got != tt.want {
				t.Errorf("FormatBlacklistForPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
