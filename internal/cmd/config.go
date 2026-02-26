package cmd

import (
	"fmt"

	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Manage AI Shell Assistant configuration. Configuration is stored in ~/.aiassist/config.yaml",
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
