package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/interactive"
	"github.com/llaoj/aiassist/internal/llm"
)

func interactiveMode() {
	cfg := config.Get()
	translator := i18n.New(cfg.GetLanguage())

	// Check if configuration file exists
	if !cfg.ConfigExists() {
		color.Red(translator.T("config.not_found") + "\n")
		fmt.Println(translator.T("config.hint_run_setup"))
		os.Exit(1)
	}

	// Check if any providers are configured
	enabledProviders := cfg.GetEnabledProviders()
	if len(enabledProviders) == 0 {
		color.Red(translator.T("error.no_models") + "\n")
		fmt.Println(translator.T("error.hint_no_models"))
		os.Exit(1)
	}

	// Initialize LLM manager
	manager := llm.NewManager(cfg)

	// Register configured providers as OpenAI-compatible providers
	// For each provider with multiple models, create separate provider instances
	for _, provider := range enabledProviders {
		for _, model := range provider.Models {
			providerKey := fmt.Sprintf("%s/%s", provider.Name, model)
			llmProvider := llm.NewOpenAICompatibleProvider(
				providerKey,
				provider.BaseURL,
				provider.APIKey,
				model,
			)
			manager.RegisterProvider(providerKey, llmProvider, provider.Priority)
		}
	}

	// Create interactive session
	session := interactive.NewSession(manager, cfg, translator, os.Stdin)

	// Check if there is pipe input
	fileInfo, _ := os.Stdin.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// Has pipe input
		var input string
		if len(os.Args) > 1 {
			input = os.Args[1]
		}

		if err := session.RunWithPipe(input); err != nil {
			color.Red("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Interactive mode
		if err := session.Run(); err != nil {
			color.Red("Error: %v\n", err)
			os.Exit(1)
		}
	}
}
