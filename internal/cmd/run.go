package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/interactive"
	"github.com/llaoj/aiassist/internal/llm"
)

// initializeSession initializes and returns an interactive session
func initializeSession() (*interactive.Session, *i18n.I18n) {
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
	for _, provider := range enabledProviders {
		for _, modelCfg := range provider.Models {
			// Skip disabled models
			if !modelCfg.Enabled {
				continue
			}

			providerKey := fmt.Sprintf("%s/%s", provider.Name, modelCfg.Name)
			llmProvider := llm.NewOpenAICompatibleProvider(
				providerKey,
				provider.BaseURL,
				provider.APIKey,
				modelCfg.Name,
			)
			manager.RegisterProvider(providerKey, llmProvider)
		}
	}

	return interactive.NewSession(manager, translator), translator
}

// runInteractiveMode runs the interactive mode
func runInteractiveMode(initialQuestion string) {
	session, translator := initializeSession()

	err := session.Run(initialQuestion)
	if err != nil {
		// Check if it's a user exit (normal termination)
		if errors.Is(err, interactive.ErrUserExit) {
			return // Normal exit, no error message
		}
		color.Red(translator.T("error.general", err) + "\n")
		os.Exit(1)
	}
}

// runPipeMode runs the pipe analysis mode
func runPipeMode(initialQuestion string) {
	session, translator := initializeSession()

	err := session.RunWithPipe(initialQuestion)
	if err != nil {
		// Check if it's a user exit (normal termination)
		if errors.Is(err, interactive.ErrUserExit) {
			return // Normal exit, no error message
		}
		color.Red(translator.T("error.general", err) + "\n")
		os.Exit(1)
	}
}
