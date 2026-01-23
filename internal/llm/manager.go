package llm

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
)

// Manager manages the lifecycle of multiple LLM providers
type Manager struct {
	providers  map[string]ModelProvider
	priority   map[string]int
	mu         sync.RWMutex
	config     *config.Config
	translator *i18n.I18n
}

// NewManager creates a new LLM manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		providers:  make(map[string]ModelProvider),
		priority:   make(map[string]int),
		config:     cfg,
		translator: i18n.New(cfg.GetLanguage()),
	}
}

// RegisterProvider registers an LLM provider
func (m *Manager) RegisterProvider(name string, provider ModelProvider, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	m.priority[name] = priority
}

// CallWithFallback calls the primary model, automatically switching to fallback models on failure
func (m *Manager) CallWithFallback(ctx context.Context, prompt string) (string, string, error) {
	return m.CallWithFallbackSystemPrompt(ctx, "", prompt)
}

// CallWithFallbackSystemPrompt calls the primary model with system prompt support, automatically switching to fallback models on failure
func (m *Manager) CallWithFallbackSystemPrompt(ctx context.Context, systemPrompt string, userPrompt string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get available providers sorted by priority
	available := m.getAvailableProviders()
	if len(available) == 0 {
		return "", "", fmt.Errorf("no available LLM providers")
	}

	var lastErr error
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
			lastErr = err
			fmt.Printf("[Warning] %s call failed, trying next model: %v\n", providerName, err)
			continue
		}

		return response, providerName, nil
	}

	return "", "", fmt.Errorf("all model calls failed: %w", lastErr)
}

// CallSpecific calls a specific model
func (m *Manager) CallSpecific(ctx context.Context, modelName string, prompt string) (string, error) {
	m.mu.RLock()
	provider, exists := m.providers[modelName]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("model %s does not exist", modelName)
	}

	if !provider.IsAvailable() {
		return "", fmt.Errorf("model %s quota exhausted or unavailable", modelName)
	}

	return provider.Call(ctx, prompt)
}

// GetAvailableProviders gets the list of available providers (sorted by priority)
func (m *Manager) getAvailableProviders() []string {
	type providerWithPriority struct {
		name     string
		priority int
	}

	var available []providerWithPriority

	for name, provider := range m.providers {
		if provider.IsAvailable() {
			available = append(available, providerWithPriority{
				name:     name,
				priority: m.priority[name],
			})
		}
	}

	// Sort by priority (descending)
	sort.Slice(available, func(i, j int) bool {
		return available[i].priority > available[j].priority
	})

	result := make([]string, len(available))
	for i, p := range available {
		result[i] = p.name
	}

	return result
}

// GetStatus gets status information for all providers
func (m *Manager) GetStatus() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]map[string]interface{})

	for name, provider := range m.providers {
		status[name] = map[string]interface{}{
			"name":            provider.GetName(),
			"available":       provider.IsAvailable(),
			"remaining_calls": provider.GetRemainingCalls(),
			"priority":        m.priority[name],
		}
	}

	return status
}

// ResetDailyQuota resets daily call quota (should be called at a specified time each day)
func (m *Manager) ResetDailyQuota() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In the new architecture, provider quota management is handled by the LLM API service
	// This method is reserved for future expansion
	return m.config.Save()
}

// PrintStatus prints the current model status to the terminal
func (m *Manager) PrintStatus() {
	status := m.GetStatus()

	fmt.Println("\n[" + m.translator.T("llm.status_title") + "]")

	for modelName, info := range status {
		available := info["available"].(bool)
		remainingCalls := info["remaining_calls"].(int)
		priority := info["priority"].(int)

		statusStr := m.translator.T("llm.status_available")
		if !available {
			statusStr = m.translator.T("llm.status_unavailable")
		}

		fmt.Printf("- %s: %s | %s: %d | %s: %d\n",
			modelName, statusStr, m.translator.T("llm.remaining_calls"), remainingCalls, m.translator.T("llm.priority"), priority)
	}

	fmt.Println()
}
