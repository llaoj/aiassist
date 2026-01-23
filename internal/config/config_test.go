package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewModelConfig(t *testing.T) {
	mc := &ModelConfig{
		Name:    "test-model",
		Enabled: true,
	}

	if mc.Name != "test-model" {
		t.Errorf("Expected Name to be 'test-model', got '%s'", mc.Name)
	}

	if !mc.Enabled {
		t.Error("Expected Enabled to be true")
	}
}

func TestProviderConfig(t *testing.T) {
	models := []*ModelConfig{
		{Name: "model1", Enabled: true},
		{Name: "model2", Enabled: false},
	}

	pc := &ProviderConfig{
		Name:    "test-provider",
		BaseURL: "https://api.test.com",
		APIKey:  "test-key",
		Models:  models,
		Enabled: true,
	}

	if pc.Name != "test-provider" {
		t.Errorf("Expected Name to be 'test-provider', got '%s'", pc.Name)
	}

	if len(pc.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(pc.Models))
	}
}

func TestConfig_GetLanguage(t *testing.T) {
	cfg := &Config{
		Language: LanguageChinese,
	}

	if cfg.GetLanguage() != LanguageChinese {
		t.Errorf("Expected language to be '%s', got '%s'", LanguageChinese, cfg.GetLanguage())
	}
}

func TestConfig_SetLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		ConfigDir:  tmpDir,
		ConfigFile: filepath.Join(tmpDir, "config.yaml"),
		Language:   LanguageEnglish,
		Providers:  make(map[string]*ProviderConfig),
	}

	err := cfg.SetLanguage(LanguageChinese)
	if err != nil {
		t.Fatalf("SetLanguage failed: %v", err)
	}

	if cfg.Language != LanguageChinese {
		t.Errorf("Expected language to be '%s', got '%s'", LanguageChinese, cfg.Language)
	}

	// Test invalid language defaults to English
	err = cfg.SetLanguage("invalid")
	if err != nil {
		t.Fatalf("SetLanguage failed: %v", err)
	}

	if cfg.Language != LanguageEnglish {
		t.Errorf("Invalid language should default to English, got '%s'", cfg.Language)
	}
}

func TestConfig_GetHTTPProxy(t *testing.T) {
	cfg := &Config{
		HTTPProxy: "http://proxy.test.com:8080",
	}

	if cfg.GetHTTPProxy() != "http://proxy.test.com:8080" {
		t.Errorf("Expected proxy to be 'http://proxy.test.com:8080', got '%s'", cfg.GetHTTPProxy())
	}
}

func TestConfig_AddProvider(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		ConfigDir:  tmpDir,
		ConfigFile: filepath.Join(tmpDir, "config.yaml"),
		Providers:  make(map[string]*ProviderConfig),
	}

	provider := &ProviderConfig{
		Name:    "test",
		BaseURL: "https://api.test.com",
		APIKey:  "key123",
		Models: []*ModelConfig{
			{Name: "model1", Enabled: true},
		},
		Enabled: true,
	}

	err := cfg.AddProvider("test", provider)
	if err != nil {
		t.Fatalf("AddProvider failed: %v", err)
	}

	retrieved := cfg.GetProvider("test")
	if retrieved == nil {
		t.Fatal("Expected provider to be added")
	}

	if retrieved.Name != "test" {
		t.Errorf("Expected provider name to be 'test', got '%s'", retrieved.Name)
	}
}

func TestConfig_GetEnabledProviders(t *testing.T) {
	cfg := &Config{
		Providers: map[string]*ProviderConfig{
			"enabled1": {
				Name:    "enabled1",
				Enabled: true,
			},
			"disabled": {
				Name:    "disabled",
				Enabled: false,
			},
			"enabled2": {
				Name:    "enabled2",
				Enabled: true,
			},
		},
	}

	enabled := cfg.GetEnabledProviders()
	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled providers, got %d", len(enabled))
	}
}

func TestConfig_ConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		ConfigFile: cfgFile,
	}

	// File doesn't exist yet
	if cfg.ConfigExists() {
		t.Error("Expected ConfigExists to return false for non-existent file")
	}

	// Create file
	if err := os.WriteFile(cfgFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File exists now
	if !cfg.ConfigExists() {
		t.Error("Expected ConfigExists to return true for existing file")
	}
}
