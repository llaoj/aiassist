package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Run interactive configuration wizard",
	Long:  "Run an interactive wizard to configure LLM providers, API keys, language preference, and proxy settings",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if Consul mode is enabled
		cfg := config.Get()
		if cfg.Consul != nil && cfg.Consul.Enabled {
			color.Red("\n✗ Configuration is managed by Consul\n")
			color.Yellow("\nPlease modify configuration using one of these methods:\n\n")
			color.Cyan("1. Consul UI:\n")
			fmt.Printf("   Visit: http://%s/ui/dc1/kv/%s/edit\n\n", cfg.Consul.Address, cfg.Consul.Key)
			color.Cyan("2. Consul CLI:\n")
			fmt.Printf("   consul kv get %s > config.yaml\n", cfg.Consul.Key)
			fmt.Printf("   # Edit config.yaml\n")
			fmt.Printf("   consul kv put %s @config.yaml\n\n", cfg.Consul.Key)
			color.Cyan("3. Consul API:\n")
			fmt.Printf("   curl http://%s/v1/kv/%s?raw > config.yaml\n", cfg.Consul.Address, cfg.Consul.Key)
			fmt.Printf("   # Edit config.yaml\n")
			fmt.Printf("   curl -X PUT --data-binary @config.yaml http://%s/v1/kv/%s\n\n", cfg.Consul.Address, cfg.Consul.Key)
			os.Exit(1) // Exit directly to avoid duplicate error message from Cobra
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return interactiveConfig()
	},
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long:  "Display current configuration details including language, proxy, default model and all providers",
	Run: func(cmd *cobra.Command, args []string) {
		viewConfig()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
}

// ensureTerminalSane ensures the terminal is in a proper state for input
// This fixes issues where liner or other tools have modified terminal settings
func ensureTerminalSane() {
	// Only run stty if stdin is a terminal
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		return // Not a TTY, skip
	}

	// Use stty to restore terminal to sane state
	if _, err := exec.LookPath("stty"); err == nil {
		cmd := exec.Command("stty", "sane")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Ignore errors - if stty fails, we'll just continue
		_ = cmd.Run()
	}
}

// readInput reads user input and handles both \r and \n line endings
func readInput(reader *bufio.Reader) string {
	var result []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		// Stop at newline (\n) or carriage return (\r)
		if b == '\n' || b == '\r' {
			// Consume any following \r or \n characters without blocking
			// Only try to read more if there's data already buffered
			for reader.Buffered() > 0 {
				next, err := reader.Peek(1)
				if err != nil {
					break
				}
				// If next char is a line ending, consume it
				if next[0] == '\n' || next[0] == '\r' {
					reader.ReadByte()
				} else {
					break
				}
			}
			break
		}
		result = append(result, b)
	}
	return string(result)
}

func viewConfig() {
	cfg := config.Get()

	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println("Current Configuration")
	fmt.Printf("%s\n\n", ui.Separator())

	// Language
	lang := cfg.GetLanguage()
	langDisplay := "English"
	if lang == config.LanguageChinese {
		langDisplay = "中文"
	}
	fmt.Printf("Language: %s (%s)\n", langDisplay, lang)

	// HTTP Proxy
	proxy := cfg.GetHTTPProxy()
	if proxy == "" {
		fmt.Printf("HTTP Proxy: Not configured\n")
	} else {
		fmt.Printf("HTTP Proxy: %s\n", proxy)
	}

	// Default Model
	defaultModel := cfg.DefaultModel
	if defaultModel == "" {
		fmt.Printf("Default Model: Not set\n")
	} else {
		fmt.Printf("Default Model: %s\n", defaultModel)
	}

	// Config file location
	fmt.Printf("Config File: %s\n", cfg.ConfigFile)

	// Providers
	allProviders := cfg.GetAllProviders()
	if len(allProviders) == 0 {
		fmt.Printf("\nProviders: None configured\n")
	} else {
		fmt.Printf("\nProviders: %d configured\n\n", len(allProviders))

		for i, p := range allProviders {
			status := "✓ Enabled"
			if !p.Enabled {
				status = "✗ Disabled"
			}
			fmt.Printf("%d. %s [%s]\n", i+1, p.Name, status)
			fmt.Printf("   Base URL: %s\n", p.BaseURL)
			fmt.Printf("   API Key: %s...%s\n", p.APIKey[:8], p.APIKey[len(p.APIKey)-4:])
			fmt.Printf("   Models:\n")

			for _, modelCfg := range p.Models {
				modelKey := fmt.Sprintf("%s/%s", p.Name, modelCfg.Name)
				modelStatus := "✓ Enabled"
				if !modelCfg.Enabled {
					modelStatus = "✗ Disabled"
				}

				// Mark default model
				defaultMark := ""
				if modelKey == defaultModel {
					defaultMark = " [DEFAULT]"
				}

				fmt.Printf("     - %s [%s]%s\n", modelKey, modelStatus, defaultMark)
			}
			fmt.Println()
		}
	}
}

func interactiveConfig() error {
	// Ensure terminal is in a sane state before starting
	// This is particularly important if liner or other tools have modified terminal settings
	ensureTerminalSane()

	reader := bufio.NewReader(os.Stdin)

	// Prompt user to select language preference first
	lang := selectLanguage(reader)
	translator := i18n.New(lang)

	// Load or create config
	cfg := config.Get()
	if err := cfg.SetLanguage(lang); err != nil {
		return fmt.Errorf("failed to set language: %w", err)
	}

	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println(translator.T("config.title"))
	fmt.Printf("%s\n\n", ui.Separator())

	// Step 1: Add OpenAI-compatible LLM providers
	fmt.Printf("\n%s\n\n", translator.T("config.openai_compat.title"))

	var addMore = true
	providerCount := 0

	for addMore {
		providerCount++
		fmt.Printf("\n%s %d:\n", translator.T("config.openai_compat.model"), providerCount)

		// Get provider name
		var providerName string
		for providerName == "" {
			fmt.Print(translator.T("config.openai_compat.provider_name"))
			name := readInput(reader)
			name = strings.TrimSpace(name)
			if name == "" {
				fmt.Println(translator.T("config.openai_compat.name_empty"))
				continue
			}
			providerName = name
		}

		// Get base URL
		var baseURL string
		for baseURL == "" {
			fmt.Print(translator.T("config.openai_compat.base_url"))
			url := readInput(reader)
			url = strings.TrimSpace(url)
			if url == "" {
				fmt.Println(translator.T("config.openai_compat.url_empty"))
				continue
			}
			baseURL = url
		}

		// Get API Key
		var apiKey string
		for apiKey == "" {
			fmt.Print(translator.T("config.openai_compat.api_key"))
			key := readInput(reader)
			key = strings.TrimSpace(key)
			if key == "" {
				fmt.Println(translator.T("config.openai_compat.key_empty"))
				continue
			}
			apiKey = key
		}

		// Get model names (comma-separated)
		var modelNames string
		for modelNames == "" {
			fmt.Print(translator.T("config.openai_compat.model_name"))
			models := readInput(reader)
			models = strings.TrimSpace(models)
			if models == "" {
				fmt.Println(translator.T("config.openai_compat.model_empty"))
				continue
			}
			modelNames = models
		}

		// Parse comma-separated model names and deduplicate
		modelList := make([]string, 0)
		seenModels := make(map[string]bool)
		for _, m := range strings.Split(modelNames, ",") {
			m = strings.TrimSpace(m)
			if m != "" && !seenModels[m] {
				modelList = append(modelList, m)
				seenModels[m] = true
			}
		}

		// Convert string list to ModelConfig list
		modelConfigs := make([]*config.ModelConfig, len(modelList))
		for i, modelName := range modelList {
			modelConfigs[i] = &config.ModelConfig{
				Name:    modelName,
				Enabled: true,
			}
		}

		// Add provider configuration
		providerCfg := &config.ProviderConfig{
			Name:    providerName,
			BaseURL: baseURL,
			APIKey:  apiKey,
			Models:  modelConfigs,
			Enabled: true,
		}
		if err := cfg.AddProvider(providerName, providerCfg); err != nil {
			color.Red(translator.T("config.error", err.Error()) + "\n")
			return err
		}
		fmt.Printf("%s\n", translator.T("config.openai_compat.success", providerName))

		// Ask if user wants to add more providers
		fmt.Print("\n" + translator.T("config.openai_compat.add_more"))
		input := readInput(reader)
		choice := strings.TrimSpace(input)
		addMore = choice == "yes" || choice == "y"
	}

	if providerCount == 0 {
		fmt.Println(translator.T("config.openai_compat.no_models"))
		return nil
	}

	// Step 2: Select default model (provider/model format)
	fmt.Printf("\n%s\n\n", translator.T("config.default_model.title"))

	enabledProviders := cfg.GetEnabledProviders()
	if len(enabledProviders) > 0 {
		// Build list of all provider/model combinations
		type modelOption struct {
			index    int
			provider string
			model    string
		}
		var allModels []modelOption
		idx := 1

		for _, p := range enabledProviders {
			for _, m := range p.Models {
				fmt.Printf("%d. %s/%s\n", idx, p.Name, m.Name)
				allModels = append(allModels, modelOption{index: idx, provider: p.Name, model: m.Name})
				idx++
			}
		}

		fmt.Print("\n" + translator.T("config.default_model.select"))
		input := readInput(reader)
		input = strings.TrimSpace(input)

		// Default to first model if empty input
		if input == "" {
			input = "1"
		}

		var selectedIdx int
		_, err := fmt.Sscanf(input, "%d", &selectedIdx)
		if err == nil && selectedIdx > 0 && selectedIdx <= len(allModels) {
			selected := allModels[selectedIdx-1]
			cfg.DefaultModel = fmt.Sprintf("%s/%s", selected.provider, selected.model)
			fmt.Printf("%s\n", translator.T("config.default_model.selected", cfg.DefaultModel))
		}
	}

	// Step 3: HTTP Proxy configuration (optional)
	fmt.Printf("\n%s\n\n", translator.T("config.proxy.title"))
	fmt.Println(translator.T("config.proxy.desc"))
	fmt.Println(translator.T("config.proxy.example"))
	fmt.Print("\n" + translator.T("config.proxy.input"))

	proxyInput := readInput(reader)
	proxyInput = strings.TrimSpace(proxyInput)

	if proxyInput != "" {
		if err := cfg.SetHTTPProxy(proxyInput); err != nil {
			fmt.Printf("%s\n", translator.T("config.proxy.error", err))
			return err
		}
		fmt.Printf("%s\n", translator.T("config.proxy.success", proxyInput))
	} else {
		if err := cfg.SetHTTPProxy(""); err != nil {
			fmt.Printf("%s\n", translator.T("config.proxy.error", err))
			return err
		}
		fmt.Println(translator.T("config.proxy.empty"))
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		fmt.Printf("%s\n", translator.T("config.save_error", err))
		return err
	}

	// Display completion message
	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println(translator.T("config.complete"))
	fmt.Printf("%s\n\n", ui.Separator())

	return nil
}

// selectLanguage prompts user to select language preference
func selectLanguage(reader *bufio.Reader) string {
	fmt.Printf("\n%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("Language Selection")
	fmt.Printf("%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("1. English")
	fmt.Println("2. Chinese (\u4e2d\u6587)")
	fmt.Print("\nPlease select (default: 1): ")

	input := readInput(reader)
	choice := strings.TrimSpace(input)

	if choice == "2" {
		return config.LanguageChinese
	}

	return config.LanguageEnglish
}
