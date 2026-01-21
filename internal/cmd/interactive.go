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

	// Check if any models are configured
	enabledModels := cfg.GetEnabledModels()
	if len(enabledModels) == 0 {
		color.Red(translator.T("error.no_models") + "\n")
		fmt.Println(translator.T("error.hint_no_models"))
		os.Exit(1)
	}

	// Initialize LLM manager
	manager := llm.NewManager(cfg)

	// Register configured models
	for _, modelCfg := range enabledModels {
		var provider llm.ModelProvider

		switch modelCfg.Name {
		case translator.T("model.qianwen"):
			provider = llm.NewQianWenProvider(modelCfg.APIKey)
		case translator.T("model.chatgpt"):
			provider = llm.NewChatGPTProvider(modelCfg.APIKey, cfg.Proxy)
		case translator.T("model.deepseek"):
			provider = llm.NewDeepSeekProvider(modelCfg.APIKey)
		default:
			color.Yellow(translator.T("error.unknown_model", modelCfg.Name) + "\n")
			continue
		}

		manager.RegisterProvider(modelCfg.Name, provider, modelCfg.Priority)
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
			color.Red("❌ Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Interactive mode
		if err := session.Run(); err != nil {
			color.Red("❌ Error: %v\n", err)
			os.Exit(1)
		}
	}
}
