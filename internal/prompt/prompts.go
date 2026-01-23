package prompt

import "github.com/llaoj/aiassist/internal/config"

// SystemPrompts defines LLM system prompts for different scenarios
type SystemPrompts struct {
	Interactive      string
	ContinueAnalysis string
	PipeAnalysis     string
}

// GetSystemPrompts returns system prompts based on language preference
func GetSystemPrompts() SystemPrompts {
	lang := config.Get().GetLanguage()

	if lang == config.LanguageChinese {
		return chinesePrompts
	}

	return englishPrompts
}

// GetInteractivePrompt returns the system prompt for interactive mode (user asks a question)
func GetInteractivePrompt() string {
	return GetSystemPrompts().Interactive
}

// GetContinueAnalysisPrompt returns the system prompt for analyzing command output from previous step
func GetContinueAnalysisPrompt() string {
	return GetSystemPrompts().ContinueAnalysis
}

// GetPipeAnalysisPrompt returns the system prompt for analyzing piped command output
func GetPipeAnalysisPrompt() string {
	return GetSystemPrompts().PipeAnalysis
}
