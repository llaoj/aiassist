package prompt

import "github.com/llaoj/aiassist/internal/config"

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
	return GetSystemPrompts().Interactive
}

func GetContinueAnalysisPrompt() string {
	return GetSystemPrompts().ContinueAnalysis
}

func GetPipeAnalysisPrompt() string {
	return GetSystemPrompts().PipeAnalysis
}
