package llm

import (
	"context"
	"testing"

	"github.com/llaoj/aiassist/internal/config"
)

// MockProvider implements ModelProvider for testing
type MockProvider struct {
	name           string
	available      bool
	remainingCalls int
	callResponse   string
	callError      error
}

func (m *MockProvider) Call(ctx context.Context, prompt string) (string, error) {
	return m.callResponse, m.callError
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) IsAvailable() bool {
	return m.available
}

func (m *MockProvider) GetRemainingCalls() int {
	return m.remainingCalls
}

func TestNewManager(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)
	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.providers == nil {
		t.Error("Expected providers map to be initialized")
	}

	if manager.providerOrder == nil {
		t.Error("Expected providerOrder to be initialized")
	}

	if manager.modelEnabled == nil {
		t.Error("Expected modelEnabled to be initialized")
	}
}

func TestRegisterProvider(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	mockProvider := &MockProvider{
		name:           "test-provider",
		available:      true,
		remainingCalls: 100,
		callResponse:   "test response",
	}

	manager.RegisterProvider("test", mockProvider)

	// Check provider was registered
	if len(manager.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(manager.providers))
	}

	if len(manager.providerOrder) != 1 {
		t.Errorf("Expected 1 provider in order, got %d", len(manager.providerOrder))
	}

	if manager.providerOrder[0] != "test" {
		t.Errorf("Expected first provider to be 'test', got '%s'", manager.providerOrder[0])
	}
}

func TestEnableDisableModel(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	// Initially enabled by default
	if !manager.isModelEnabled("test-model") {
		t.Error("Expected model to be enabled by default")
	}

	// Disable model
	manager.DisableModel("test-model")
	if manager.isModelEnabled("test-model") {
		t.Error("Expected model to be disabled")
	}

	// Enable model
	manager.EnableModel("test-model")
	if !manager.isModelEnabled("test-model") {
		t.Error("Expected model to be enabled")
	}
}

func TestGetAvailableProviders_Empty(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	available := manager.getAvailableProviders()
	if len(available) != 0 {
		t.Errorf("Expected 0 available providers, got %d", len(available))
	}
}

func TestGetAvailableProviders_WithProviders(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	// Register providers in order
	provider1 := &MockProvider{name: "provider1", available: true, remainingCalls: 100}
	provider2 := &MockProvider{name: "provider2", available: true, remainingCalls: 50}
	provider3 := &MockProvider{name: "provider3", available: false, remainingCalls: 0}

	manager.RegisterProvider("provider1", provider1)
	manager.RegisterProvider("provider2", provider2)
	manager.RegisterProvider("provider3", provider3)

	available := manager.getAvailableProviders()

	// Should return 2 available providers (provider3 is unavailable)
	if len(available) != 2 {
		t.Fatalf("Expected 2 available providers, got %d", len(available))
	}

	// Should be in registration order
	if available[0] != "provider1" {
		t.Errorf("Expected first provider to be 'provider1', got '%s'", available[0])
	}

	if available[1] != "provider2" {
		t.Errorf("Expected second provider to be 'provider2', got '%s'", available[1])
	}
}

func TestGetAvailableProviders_RespectsDisabledModels(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	provider := &MockProvider{name: "provider", available: true, remainingCalls: 100}
	manager.RegisterProvider("test-model", provider)

	// Model is available
	available := manager.getAvailableProviders()
	if len(available) != 1 {
		t.Fatalf("Expected 1 available provider, got %d", len(available))
	}

	// Disable the model
	manager.DisableModel("test-model")
	available = manager.getAvailableProviders()
	if len(available) != 0 {
		t.Errorf("Expected 0 available providers after disabling, got %d", len(available))
	}
}

func TestGetStatus(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	provider := &MockProvider{
		name:           "test-provider",
		available:      true,
		remainingCalls: 42,
	}

	manager.RegisterProvider("test", provider)

	status := manager.GetStatus()

	if len(status) != 1 {
		t.Fatalf("Expected 1 status entry, got %d", len(status))
	}

	testStatus, exists := status["test"]
	if !exists {
		t.Fatal("Expected status for 'test' provider")
	}

	if testStatus["name"] != "test-provider" {
		t.Errorf("Expected name to be 'test-provider', got '%v'", testStatus["name"])
	}

	if testStatus["available"] != true {
		t.Errorf("Expected available to be true, got %v", testStatus["available"])
	}

	if testStatus["remaining_calls"] != 42 {
		t.Errorf("Expected remaining_calls to be 42, got %v", testStatus["remaining_calls"])
	}

	if testStatus["enabled"] != true {
		t.Errorf("Expected enabled to be true, got %v", testStatus["enabled"])
	}
}

func TestCallWithFallback_NoProviders(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)
	ctx := context.Background()

	_, _, err := manager.CallWithFallback(ctx, "test prompt")
	if err == nil {
		t.Error("Expected error when no providers available")
	}
}

func TestCallWithFallback_Success(t *testing.T) {
	cfg := &config.Config{
		Language:  config.LanguageEnglish,
		Providers: make(map[string]*config.ProviderConfig),
	}

	manager := NewManager(cfg)

	provider := &MockProvider{
		name:           "test-provider",
		available:      true,
		remainingCalls: 100,
		callResponse:   "mock response",
		callError:      nil,
	}

	manager.RegisterProvider("test", provider)

	ctx := context.Background()
	response, modelUsed, err := manager.CallWithFallback(ctx, "test prompt")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response != "mock response" {
		t.Errorf("Expected response 'mock response', got '%s'", response)
	}

	if modelUsed != "test" {
		t.Errorf("Expected modelUsed to be 'test', got '%s'", modelUsed)
	}
}
