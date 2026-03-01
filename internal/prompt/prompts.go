package prompt

import (
	"strings"

	"github.com/llaoj/aiassist/internal/blacklist"
	"github.com/llaoj/aiassist/internal/config"
)

// SystemPrompts defines LLM system prompts for different scenarios
type SystemPrompts struct {
	Interactive      string
	ContinueAnalysis string
	PipeAnalysis     string
}

func GetSystemPrompts() SystemPrompts {
	lang := config.Get().GetLanguage()

	if lang == config.LanguageChinese {
		return chinesePrompts
	}

	return englishPrompts
}

func GetInteractivePrompt() string {
	prompt := GetSystemPrompts().Interactive
	return injectBlacklist(prompt)
}

func GetContinueAnalysisPrompt() string {
	prompt := GetSystemPrompts().ContinueAnalysis
	return injectBlacklist(prompt)
}

func GetPipeAnalysisPrompt() string {
	prompt := GetSystemPrompts().PipeAnalysis
	return injectBlacklist(prompt)
}

// injectBlacklist replaces {{COMMAND_BLACKLIST}} placeholder with actual blacklist content
func injectBlacklist(prompt string) string {
	checker := blacklist.NewChecker()
	blacklistText := checker.FormatBlacklistForPrompt()

	// If blacklist is empty, replace with empty indication
	if blacklistText == "" {
		// Remove the [Command Blacklist]: section entirely
		// The placeholder line is: [Command Blacklist]:\n{{COMMAND_BLACKLIST}}
		return strings.ReplaceAll(prompt, "{{COMMAND_BLACKLIST}}", "无（未配置黑名单）")
	}

	// Replace placeholder with actual blacklist
	return strings.ReplaceAll(prompt, "{{COMMAND_BLACKLIST}}", blacklistText)
}

