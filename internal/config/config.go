package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Language constants
const (
	LanguageEnglish = "en"
	LanguageChinese = "zh"
)

// ModelConfig represents model configuration
type ModelConfig struct {
	Name          string `yaml:"name"`
	APIKey        string `yaml:"api_key"`
	Priority      int    `yaml:"priority"`
	MaxCalls      int    `yaml:"max_calls_per_day"`
	CurrentCalls  int    `yaml:"current_calls"`
	LastResetTime int64  `yaml:"last_reset_time"`
	Enabled       bool   `yaml:"enabled"`
}

// Config represents global configuration
type Config struct {
	Language       string                  `yaml:"language"`
	Proxy          string                  `yaml:"proxy"`
	MaxConcurrency int                     `yaml:"max_concurrency"`
	DefaultModel   string                  `yaml:"default_model"`
	DailyResetHour int                     `yaml:"daily_reset_hour"`
	Models         map[string]*ModelConfig `yaml:"models"`

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
		Language:       LanguageEnglish,
		MaxConcurrency: 5,
		DailyResetHour: 0,
		Proxy:          "",
		DefaultModel:   "",
		Models:         make(map[string]*ModelConfig),
		ConfigDir:      configDir,
		ConfigFile:     configFile,
	}

	// Load config file if it exists
	if _, err := os.Stat(configFile); err == nil {
		return globalConfig.Load()
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

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(c.ConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddModel adds a model configuration
func (c *Config) AddModel(name string, modelConfig *ModelConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Models == nil {
		c.Models = make(map[string]*ModelConfig)
	}

	c.Models[name] = modelConfig
	return c.Save()
}

// GetModel gets a model configuration
func (c *Config) GetModel(name string) *ModelConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Models[name]
}

// GetEnabledModels returns list of enabled models
func (c *Config) GetEnabledModels() []*ModelConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	enabled := make([]*ModelConfig, 0)
	for _, model := range c.Models {
		if model.Enabled {
			enabled = append(enabled, model)
		}
	}

	return enabled
}

// UpdateModelCalls updates model call count
func (c *Config) UpdateModelCalls(modelName string, increment bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	model, ok := c.Models[modelName]
	if !ok {
		return fmt.Errorf("model %s not found", modelName)
	}

	if increment {
		model.CurrentCalls++
	} else {
		if model.CurrentCalls > 0 {
			model.CurrentCalls--
		}
	}

	return c.Save()
}

// SetProxy sets the global proxy
func (c *Config) SetProxy(proxy string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Proxy = proxy
	return c.Save()
}

// GetProxy returns the configured proxy address
func (c *Config) GetProxy() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Proxy
}

// SetLanguage sets the language preference
func (c *Config) SetLanguage(lang string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lang != LanguageEnglish && lang != LanguageChinese {
		lang = LanguageEnglish
	}

	c.Language = lang
	return c.Save()
}

// GetLanguage returns the language preference
func (c *Config) GetLanguage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Language
}

// ValidateAPIKey validates API Key
func (c *Config) ValidateAPIKey(modelName, apiKey string) (bool, error) {
	// This should implement actual API Key validation logic
	// Send a lightweight test request to the model's API to verify the Key
	return true, nil
}

// ConfigExists checks if configuration file exists
func (c *Config) ConfigExists() bool {
	if c.ConfigFile == "" {
		return false
	}

	_, err := os.Stat(c.ConfigFile)
	return err == nil
}
