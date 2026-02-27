package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
)

// Manager manages the lifecycle of multiple LLM providers
type Manager struct {
	providers     map[string]ModelProvider
	providerOrder []string // Order of providers from config file
	mu            sync.RWMutex
	config        *config.Config
	translator    *i18n.I18n
	modelEnabled  map[string]bool // Track enabled status for each model
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		providers:     make(map[string]ModelProvider),
		providerOrder: make([]string, 0),
		modelEnabled:  make(map[string]bool),
		config:        cfg,
		translator:    i18n.New(cfg.GetLanguage()),
	}
}

func (m *Manager) RegisterProvider(name string, provider ModelProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	// Add to order if not already present
	found := false
	for _, n := range m.providerOrder {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		m.providerOrder = append(m.providerOrder, name)
	}
	// Initialize model as enabled
	m.modelEnabled[name] = true
}

func (m *Manager) CallWithFallback(ctx context.Context, prompt string) (string, string, error) {
	return m.CallWithFallbackSystemPrompt(ctx, "", prompt)
}

func (m *Manager) CallWithFallbackSystemPrompt(ctx context.Context, systemPrompt string, userPrompt string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get available providers sorted by priority
	available := m.getAvailableProviders()
	if len(available) == 0 {
		return "", "", fmt.Errorf("no available LLM providers")
	}

	for _, providerName := range available {
		provider := m.providers[providerName]

		// Check timeout context
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		default:
		}

		// Attempt to call
		var response string
		var err error

		// If provider supports system prompt, use the version with system prompt
		if compatProvider, ok := provider.(*OpenAICompatibleProvider); ok && systemPrompt != "" {
			response, err = compatProvider.CallWithSystemPrompt(ctx, systemPrompt, userPrompt)
		} else {
			response, err = provider.Call(ctx, userPrompt)
		}

		if err != nil {
			color.Red("Error: %s call failed: %v\n", providerName, err)
			continue
		}

		return response, providerName, nil
	}

	return "", "", fmt.Errorf("all model calls failed")
}

func (m *Manager) getAvailableProviders() []string {
	var available []string

	for _, name := range m.providerOrder {
		provider, exists := m.providers[name]
		if !exists {
			continue
		}
		if provider.IsAvailable() && m.isModelEnabled(name) {
			available = append(available, name)
		}
	}

	return available
}

func (m *Manager) isModelEnabled(modelName string) bool {
	enabled, exists := m.modelEnabled[modelName]
	if !exists {
		return true // Default to enabled if not set
	}
	return enabled
}

func (m *Manager) DisableModel(modelName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelEnabled[modelName] = false
}

func (m *Manager) EnableModel(modelName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelEnabled[modelName] = true
}

func (m *Manager) GetStatus() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]map[string]interface{})

	for name, provider := range m.providers {
		isAvailable := provider.IsAvailable() && m.isModelEnabled(name)
		status[name] = map[string]interface{}{
			"name":            provider.GetName(),
			"available":       isAvailable,
			"remaining_calls": provider.GetRemainingCalls(),
			"enabled":         m.isModelEnabled(name),
		}
	}

	return status
}

func (m *Manager) PrintStatus() {
	status := m.GetStatus()

	fmt.Println("\n[" + m.translator.T("llm.status_title") + "]")

	for modelName, info := range status {
		available := info["available"].(bool)

		statusStr := m.translator.T("llm.status_available")
		if !available {
			statusStr = m.translator.T("llm.status_unavailable")
		}

		fmt.Printf("- %s: %s\n", modelName, statusStr)
	}
}
