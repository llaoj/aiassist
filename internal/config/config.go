package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Language constants
const (
	LanguageEnglish = "en"
	LanguageChinese = "zh"
)

// ModelConfig represents a single model configuration
type ModelConfig struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
}

// ProviderConfig represents a single LLM provider configuration
type ProviderConfig struct {
	Name    string         `yaml:"name"`
	BaseURL string         `yaml:"base_url"`
	APIKey  string         `yaml:"api_key"`
	Models  []*ModelConfig `yaml:"models"`
	Enabled bool           `yaml:"enabled"`
}

// ConsulConfig represents Consul configuration center settings
type ConsulConfig struct {
	Enabled bool   `yaml:"enabled"`         // Enable Consul config center
	Address string `yaml:"address"`         // Consul address (e.g., "127.0.0.1:8500")
	Key     string `yaml:"key"`             // KV key to store config (e.g., "aiassist/config")
	Token   string `yaml:"token,omitempty"` // ACL token (optional)
}

// Config represents global configuration
type Config struct {
	Language     string                     `yaml:"language"`
	DefaultModel string                     `yaml:"default_model"`
	Consul       *ConsulConfig              `yaml:"consul,omitempty"` // Consul config center settings
	Providers    map[string]*ProviderConfig `yaml:"providers"`

	ConfigDir  string       `yaml:"-"`
	ConfigFile string       `yaml:"-"`
	mu         sync.RWMutex `yaml:"-"`
}

var globalConfig *Config

// Init initializes global configuration
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".aiassist")

	// Create configuration directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Initialize config structure
	globalConfig = &Config{
		Language:     LanguageEnglish,
		DefaultModel: "",
		Providers:    make(map[string]*ProviderConfig),
		ConfigDir:    configDir,
		ConfigFile:   configFile,
	}

	// Load local config file if it exists
	if _, err := os.Stat(configFile); err == nil {
		if err := globalConfig.Load(); err != nil {
			return err
		}

		// Check if Consul is configured and enabled
		if globalConfig.Consul != nil && globalConfig.Consul.Enabled {
			// Try to load from Consul
			cfg, err := LoadFromConsul(globalConfig.Consul)
			if err == nil {
				// Successfully loaded from Consul, use those providers
				globalConfig.Language = cfg.Language
				globalConfig.DefaultModel = cfg.DefaultModel
				globalConfig.Providers = cfg.Providers
				return nil
			}
			// Consul load failed, continue using local providers
		}

		return nil
	}

	// Config file doesn't exist - return nil
	// Caller will check ConfigExists() and prompt user to run: aiassist config
	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		Init()
	}
	return globalConfig
}

// Load loads configuration from file
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.save()
}

// save saves configuration to file (internal, caller must hold lock)
// Note: Consul mode check is handled at cmd layer via PersistentPreRunE
func (c *Config) save() error {
	return c.saveToFile()
}

// saveToFile saves configuration to local file
func (c *Config) saveToFile() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(c.ConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddProvider adds a provider configuration
func (c *Config) AddProvider(providerName string, provider *ProviderConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Providers == nil {
		c.Providers = make(map[string]*ProviderConfig)
	}

	c.Providers[providerName] = provider
	return c.save()
}

// GetProvider gets a provider configuration
func (c *Config) GetProvider(name string) *ProviderConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Providers[name]
}

// GetEnabledProviders returns list of enabled providers
func (c *Config) GetEnabledProviders() []*ProviderConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	enabled := make([]*ProviderConfig, 0)
	for _, provider := range c.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}

	return enabled
}

// GetAllProviders returns list of all providers
func (c *Config) GetAllProviders() []*ProviderConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make([]*ProviderConfig, 0, len(c.Providers))
	for _, provider := range c.Providers {
		all = append(all, provider)
	}

	return all
}

// DeleteProvider removes a provider configuration
func (c *Config) DeleteProvider(providerName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Providers == nil {
		return fmt.Errorf("no providers configured")
	}

	if _, exists := c.Providers[providerName]; !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	// If this was the default model, clear it
	if strings.HasPrefix(c.DefaultModel, providerName+"/") {
		c.DefaultModel = ""
	}

	delete(c.Providers, providerName)
	return c.save()
}

// SetLanguage sets the language preference
func (c *Config) SetLanguage(lang string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lang != LanguageEnglish && lang != LanguageChinese {
		lang = LanguageEnglish
	}

	c.Language = lang
	return c.save()
}

// GetLanguage returns the language preference
func (c *Config) GetLanguage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Language
}

// ConfigExists checks if configuration exists (either in Consul or local file)
func (c *Config) ConfigExists() bool {
	// If using Consul mode, check if providers exist
	if c.Consul != nil && c.Consul.Enabled {
		c.mu.RLock()
		hasProviders := len(c.Providers) > 0
		c.mu.RUnlock()
		return hasProviders
	}

	// Otherwise check local file
	if c.ConfigFile == "" {
		return false
	}

	_, err := os.Stat(c.ConfigFile)
	return err == nil
}
