package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Interactive configuration wizard",
	Long:  "Configure LLM models, API Keys, language preference, proxy and other parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return interactiveConfig()
	},
}

func interactiveConfig() error {
	reader := bufio.NewReader(os.Stdin)

	// Prompt user to select language preference first
	lang := selectLanguage(reader)
	translator := i18n.New(lang)

	// Load or create config
	cfg := config.Get()
	cfg.SetLanguage(lang)

	fmt.Printf("\n%s\n", translator.T("config.separator"))
	fmt.Println(translator.T("config.title"))
	fmt.Printf("%s\n\n", translator.T("config.separator"))

	// Step 1: Model selection
	fmt.Printf("%s\n\n", translator.T("config.step1.title"))

	models := []struct {
		id       int
		key      string
		descKey  string
		urlKey   string
		guideKey string
	}{
		{
			id:       1,
			key:      "model.qianwen",
			descKey:  "model.qianwen.desc",
			urlKey:   "model.qianwen.url",
			guideKey: "model.qianwen.guide",
		},
		{
			id:       2,
			key:      "model.chatgpt",
			descKey:  "model.chatgpt.desc",
			urlKey:   "model.chatgpt.url",
			guideKey: "model.chatgpt.guide",
		},
		{
			id:       3,
			key:      "model.deepseek",
			descKey:  "model.deepseek.desc",
			urlKey:   "model.deepseek.url",
			guideKey: "model.deepseek.guide",
		},
	}

	// Display available models
	for _, model := range models {
		fmt.Printf("%d. %s (%s)\n",
			model.id,
			translator.T(model.key),
			translator.T(model.descKey))
	}

	// Read user input for model selection
	fmt.Print("\n" + translator.T("config.step1.input"))
	input, _ := reader.ReadString('\n')
	selectedIDs := strings.Split(strings.TrimSpace(input), ",")

	selectedModels := make([]int, 0)
	for _, id := range selectedIDs {
		var num int
		_, err := fmt.Sscanf(strings.TrimSpace(id), "%d", &num)
		if err == nil && num > 0 && num <= len(models) {
			selectedModels = append(selectedModels, num)
		}
	}

	if len(selectedModels) == 0 {
		fmt.Println(translator.T("config.step1.error"))
		return nil
	}

	// Step 2: Configure selected models
	fmt.Printf("\n%s\n\n", translator.T("config.step2.title"))

	for _, modelID := range selectedModels {
		for _, model := range models {
			if model.id == modelID {
				modelName := translator.T(model.key)
				fmt.Printf("\n%s\n", translator.T("config.step2.config", modelName))
				fmt.Printf("%s\n", translator.T("config.step2.api_url", translator.T(model.urlKey)))
				fmt.Printf("%s\n\n", translator.T("config.step2.guide", translator.T(model.guideKey)))
				fmt.Print(translator.T("config.step2.api_key_input"))

				apiKey, _ := reader.ReadString('\n')
				apiKey = strings.TrimSpace(apiKey)

				if apiKey != "" {
					fmt.Println(translator.T("config.step2.validating"))
					if valid, _ := cfg.ValidateAPIKey(modelName, apiKey); valid {
						fmt.Println(translator.T("config.step2.success"))

						// Add model configuration
						modelCfg := &config.ModelConfig{
							Name:         modelName,
							APIKey:       apiKey,
							Priority:     modelID,
							MaxCalls:     1000,
							CurrentCalls: 0,
							Enabled:      true,
						}
						cfg.AddModel(modelName, modelCfg)
					} else {
						fmt.Println(translator.T("config.step2.failed"))
					}
				} else {
					fmt.Println(translator.T("config.step2.empty"))
				}

				break
			}
		}
	}

	// Step 3: Proxy configuration
	fmt.Printf("\n%s\n\n", translator.T("config.step3.title"))
	fmt.Println(translator.T("config.step3.desc"))
	fmt.Println(translator.T("config.step3.example"))
	fmt.Print("\n" + translator.T("config.step3.input"))

	proxyInput, _ := reader.ReadString('\n')
	proxyInput = strings.TrimSpace(proxyInput)

	if proxyInput != "" {
		if err := cfg.SetProxy(proxyInput); err != nil {
			fmt.Printf("%s\n", translator.T("config.step3.error", err))
			return err
		}
		fmt.Printf("%s\n", translator.T("config.step3.success", proxyInput))
	} else {
		if err := cfg.SetProxy(""); err != nil {
			fmt.Printf("%s\n", translator.T("config.step3.error", err))
			return err
		}
		fmt.Println(translator.T("config.step3.empty"))
	}

	// Step 4: Call limits configuration
	fmt.Printf("\n%s\n", translator.T("config.step4.title"))
	fmt.Print(translator.T("config.step4.input"))

	limitInput, _ := reader.ReadString('\n')
	limit := strings.TrimSpace(limitInput)
	if limit == "" {
		limit = "1000"
	}
	fmt.Printf("%s\n", translator.T("config.step4.success", limit))

	// Display completion message
	fmt.Printf("\n%s\n", translator.T("config.separator"))
	fmt.Println(translator.T("config.complete"))
	fmt.Printf("%s\n\n", translator.T("config.separator"))

	return nil
}

// selectLanguage prompts user to select language preference
func selectLanguage(reader *bufio.Reader) string {
	fmt.Printf("\n%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("Language Selection / 语言选择")
	fmt.Printf("%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("1. English")
	fmt.Println("2. 中文")
	fmt.Print("\nPlease select / 请选择 (default: 1): ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "2" {
		return config.LanguageChinese
	}

	return config.LanguageEnglish
}
