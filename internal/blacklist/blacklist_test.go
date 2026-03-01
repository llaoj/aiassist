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
		// --- Basic wildcard patterns ---
		{
			name:      "rm * matches rm with args",
			blacklist: []string{"rm *"},
			command:   "rm -rf /",
			want:      true,
			pattern:   "rm *",
		},
		{
			name:      "rm * matches rm with single file",
			blacklist: []string{"rm *"},
			command:   "rm file.txt",
			want:      true,
			pattern:   "rm *",
		},
		{
			name:      "rm * does NOT match rm with no args (wildcard requires at least the prefix words)",
			blacklist: []string{"rm *"},
			command:   "rm",
			want:      false,
			pattern:   "",
		},
		{
			name:      "dd * matches dd with args",
			blacklist: []string{"dd *"},
			command:   "dd if=/dev/zero of=/dev/sda",
			want:      true,
			pattern:   "dd *",
		},

		// --- Multi-word patterns ---
		{
			name:      "kubectl delete * matches kubectl delete pod",
			blacklist: []string{"kubectl delete *"},
			command:   "kubectl delete pod nginx",
			want:      true,
			pattern:   "kubectl delete *",
		},
		{
			name:      "kubectl delete * does NOT match kubectl get",
			blacklist: []string{"kubectl delete *"},
			command:   "kubectl get pods",
			want:      false,
			pattern:   "",
		},
		{
			name:      "kubectl delete * does NOT match kubectl alone",
			blacklist: []string{"kubectl delete *"},
			command:   "kubectl",
			want:      false,
			pattern:   "",
		},

		// --- Exact patterns (no wildcard) ---
		{
			name:      "exact match shutdown",
			blacklist: []string{"shutdown"},
			command:   "shutdown",
			want:      true,
			pattern:   "shutdown",
		},
		{
			name:      "exact pattern also matches with trailing args",
			blacklist: []string{"shutdown"},
			command:   "shutdown -h now",
			want:      true,
			pattern:   "shutdown",
		},
		{
			name:      "exact multi-word pattern matches",
			blacklist: []string{"rm -rf"},
			command:   "rm -rf /tmp/foo",
			want:      true,
			pattern:   "rm -rf",
		},
		{
			name:      "exact multi-word pattern does not match partial",
			blacklist: []string{"rm -rf"},
			command:   "rm -r /tmp",
			want:      false,
			pattern:   "",
		},

		// --- Base name normalization ---
		{
			name:      "absolute path command matches base-name pattern",
			blacklist: []string{"rm *"},
			command:   "/usr/bin/rm -rf /",
			want:      true,
			pattern:   "rm *",
		},
		{
			name:      "absolute path in pattern matches plain command",
			blacklist: []string{"/usr/bin/rm *"},
			command:   "rm -rf /",
			want:      true,
			pattern:   "/usr/bin/rm *",
		},

		// --- No match cases ---
		{
			name:      "safe command not matched",
			blacklist: []string{"rm *"},
			command:   "ls -la",
			want:      false,
			pattern:   "",
		},
		{
			name:      "command sharing prefix but different word not matched",
			blacklist: []string{"rm *"},
			command:   "remove-dir foo",
			want:      false,
			pattern:   "",
		},

		// --- Edge cases ---
		{
			name:      "empty blacklist",
			blacklist: []string{},
			command:   "rm -rf /",
			want:      false,
			pattern:   "",
		},
		{
			name:      "empty command",
			blacklist: []string{"rm *"},
			command:   "",
			want:      false,
			pattern:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &Checker{blacklist: tt.blacklist}
			got, pattern := checker.IsBlacklisted(tt.command)
			if got != tt.want {
				t.Errorf("IsBlacklisted(%q) = %v, want %v", tt.command, got, tt.want)
			}
			if got && pattern != tt.pattern {
				t.Errorf("IsBlacklisted(%q) pattern = %q, want %q", tt.command, pattern, tt.pattern)
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
