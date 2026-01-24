package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/llm"
	"github.com/llaoj/aiassist/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Configure LLM providers, API keys, language preference, proxy and other parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand provided, run interactive configuration wizard
		if len(args) == 0 {
			return interactiveConfig()
		}
		return nil
	},
}

var configProviderCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage LLM providers",
	Long:  "Add, delete, enable, disable, or list LLM providers",
}

var configProviderAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new LLM provider",
	Long:  "Add a new LLM provider with base URL, API Key, and model names",
	Run: func(cmd *cobra.Command, args []string) {
		addProviderInteractive()
	},
}

var configProviderDeleteCmd = &cobra.Command{
	Use:   "delete [provider-name]",
	Short: "Delete an LLM provider",
	Long:  "Delete an LLM provider from the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			deleteProviderByName(args[0])
		} else {
			deleteProviderInteractive()
		}
	},
}

var configProviderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all LLM providers",
	Long:  "List all configured LLM providers",
	Run: func(cmd *cobra.Command, args []string) {
		listProviders()
	},
}

var configProviderEnableCmd = &cobra.Command{
	Use:   "enable [provider-name]",
	Short: "Enable an LLM provider",
	Long:  "Enable a disabled LLM provider",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Please specify provider name\n")
			return
		}
		enableProvider(args[0])
	},
}

var configProviderDisableCmd = &cobra.Command{
	Use:   "disable [provider-name]",
	Short: "Disable an LLM provider",
	Long:  "Disable an LLM provider without deleting it",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Please specify provider name\n")
			return
		}
		disableProvider(args[0])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configProviderCmd)
	configProviderCmd.AddCommand(configProviderAddCmd)
	configProviderCmd.AddCommand(configProviderDeleteCmd)
	configProviderCmd.AddCommand(configProviderListCmd)
	configProviderCmd.AddCommand(configProviderEnableCmd)
	configProviderCmd.AddCommand(configProviderDisableCmd)
}

func addProviderInteractive() {
	reader := bufio.NewReader(os.Stdin)

	// Load config
	cfg := config.Get()
	translator := i18n.New(cfg.GetLanguage())

	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println(translator.T("config.openai_compat.title"))
	fmt.Printf("%s\n\n", ui.Separator())

	// Get provider name
	var providerName string
	for providerName == "" {
		fmt.Print(translator.T("config.openai_compat.provider_name"))
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			fmt.Println(translator.T("config.openai_compat.name_empty"))
			continue
		}

		// Check if provider already exists
		if cfg.GetProvider(name) != nil {
			color.Red(fmt.Sprintf("Provider '%s' already exists\n", name))
			continue
		}

		providerName = name
	}

	// Get base URL
	var baseURL string
	for baseURL == "" {
		fmt.Print(translator.T("config.openai_compat.base_url"))
		url, _ := reader.ReadString('\n')
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
		key, _ := reader.ReadString('\n')
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
		models, _ := reader.ReadString('\n')
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
		color.Red(fmt.Sprintf("Failed to add provider: %v\n", err))
		return
	}

	color.Green(translator.T("config.openai_compat.added", providerName) + "\n")
	fmt.Printf(translator.T("config.openai_compat.models_list", modelList) + "\n")
	color.Yellow("\n" + translator.T("config.openai_compat.order_hint") + "\n\n")
}

func deleteProviderInteractive() {
	reader := bufio.NewReader(os.Stdin)

	// Load config
	cfg := config.Get()

	allProviders := cfg.GetEnabledProviders()
	if len(allProviders) == 0 {
		color.Yellow("⚠ No providers configured\n")
		return
	}

	// List providers
	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println("Available Providers:")
	fmt.Printf("%s\n\n", ui.Separator())

	for i, p := range allProviders {
		fmt.Printf("%d. %s\n", i+1, p.Name)
		fmt.Printf("   Models: %v\n", p.Models)
	}

	// Ask user to select provider
	fmt.Print("\nEnter provider number to delete (or press Enter to cancel): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("Cancelled")
		return
	}

	var idx int
	_, err := fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > len(allProviders) {
		color.Red("Invalid selection\n")
		return
	}

	providerName := allProviders[idx-1].Name

	// Confirm deletion
	fmt.Printf("\nAre you sure you want to delete '%s'? (yes/no): ", providerName)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "yes" && confirm != "y" {
		fmt.Println("Cancelled")
		return
	}

	deleteProviderByName(providerName)
}

func deleteProviderByName(providerName string) {
	cfg := config.Get()

	if err := cfg.DeleteProvider(providerName); err != nil {
		color.Red(fmt.Sprintf("Failed to delete provider: %v\n", err))
		return
	}

	color.Green(fmt.Sprintf("✓ Provider '%s' deleted successfully\n", providerName))
}

func listProviders() {
	cfg := config.Get()

	allProviders := cfg.GetAllProviders()
	if len(allProviders) == 0 {
		color.Yellow("⚠ No providers configured\n")
		return
	}

	// Create LLM manager to check model availability
	manager := llm.NewManager(cfg)

	// Register all providers/models to get their status
	modelStatusMap := make(map[string]map[string]interface{})
	seenModels := make(map[string]bool)

	for _, provider := range allProviders {
		for _, modelCfg := range provider.Models {
			if !modelCfg.Enabled {
				continue
			}

			modelKey := fmt.Sprintf("%s/%s", provider.Name, modelCfg.Name)

			// Skip duplicate models
			if seenModels[modelKey] {
				continue
			}
			seenModels[modelKey] = true

			// Create provider instance to check status
			p := llm.NewOpenAICompatibleProvider(
				modelKey,
				provider.BaseURL,
				provider.APIKey,
				modelCfg.Name,
			)
			manager.RegisterProvider(modelKey, p)
		}
	}

	// Get status for all registered models
	modelStatusMap = manager.GetStatus()

	fmt.Printf("\n%s\n", ui.Separator())
	fmt.Println("Configured Providers:")
	fmt.Printf("%s\n\n", ui.Separator())

	for i, p := range allProviders {
		status := "✓ Enabled"
		if !p.Enabled {
			status = "✗ Disabled"
		}
		fmt.Printf("%d. %s [%s]\n", i+1, p.Name, status)
		fmt.Printf("   Base URL: %s\n", p.BaseURL)

		// Display models with their status (deduplicated)
		fmt.Printf("   Models:\n")
		displayedModels := make(map[string]bool)

		for _, modelCfg := range p.Models {
			modelKey := fmt.Sprintf("%s/%s", p.Name, modelCfg.Name)

			// Skip duplicates
			if displayedModels[modelKey] {
				continue
			}
			displayedModels[modelKey] = true

			// Get model status
			modelStatus := "✓ Enabled"
			if !modelCfg.Enabled {
				modelStatus = "✗ Disabled"
			} else if statusInfo, exists := modelStatusMap[modelKey]; exists {
				if available, ok := statusInfo["available"].(bool); ok && !available {
					modelStatus = "✗ Unavailable"
				}
			}

			fmt.Printf("     - %s [%s]\n", modelCfg.Name, modelStatus)
		}
		fmt.Println()
	}
}

func enableProvider(providerName string) {
	cfg := config.Get()
	provider := cfg.GetProvider(providerName)

	if provider == nil {
		color.Red(fmt.Sprintf("Provider '%s' not found\n", providerName))
		return
	}

	if provider.Enabled {
		color.Yellow(fmt.Sprintf("⚠ Provider '%s' is already enabled\n", providerName))
		return
	}

	provider.Enabled = true
	if err := cfg.AddProvider(providerName, provider); err != nil {
		color.Red(fmt.Sprintf("Failed to enable provider: %v\n", err))
		return
	}

	color.Green(fmt.Sprintf("✓ Provider '%s' enabled successfully\n", providerName))
}

func disableProvider(providerName string) {
	cfg := config.Get()
	provider := cfg.GetProvider(providerName)

	if provider == nil {
		color.Red(fmt.Sprintf("Provider '%s' not found\n", providerName))
		return
	}

	if !provider.Enabled {
		color.Yellow(fmt.Sprintf("⚠ Provider '%s' is already disabled\n", providerName))
		return
	}

	provider.Enabled = false
	if err := cfg.AddProvider(providerName, provider); err != nil {
		color.Red(fmt.Sprintf("Failed to disable provider: %v\n", err))
		return
	}

	color.Green(fmt.Sprintf("✓ Provider '%s' disabled successfully\n", providerName))
}

func interactiveConfig() error {
	reader := bufio.NewReader(os.Stdin)

	// Prompt user to select language preference first
	lang := selectLanguage(reader)
	translator := i18n.New(lang)

	// Load or create config
	cfg := config.Get()
	cfg.SetLanguage(lang)

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
			name, _ := reader.ReadString('\n')
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
			url, _ := reader.ReadString('\n')
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
			key, _ := reader.ReadString('\n')
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
			models, _ := reader.ReadString('\n')
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
		cfg.AddProvider(providerName, providerCfg)
		fmt.Printf("%s\n", translator.T("config.openai_compat.success", providerName))

		// Ask if user wants to add more providers
		fmt.Print("\n" + translator.T("config.openai_compat.add_more"))
		input, _ := reader.ReadString('\n')
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
		input, _ := reader.ReadString('\n')
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

	proxyInput, _ := reader.ReadString('\n')
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

// getModelNames extracts model names from ModelConfig list
func getModelNames(models []*config.ModelConfig) []string {
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.Name
	}
	return names
}

// selectLanguage prompts user to select language preference
func selectLanguage(reader *bufio.Reader) string {
	fmt.Printf("\n%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("Language Selection")
	fmt.Printf("%s\n\n", "═══════════════════════════════════════════")
	fmt.Println("1. English")
	fmt.Println("2. Chinese (\u4e2d\u6587)")
	fmt.Print("\nPlease select (default: 1): ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	if choice == "2" {
		return config.LanguageChinese
	}

	return config.LanguageEnglish
}
