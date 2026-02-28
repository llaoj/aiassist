package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/ui"
)

// Manager manages the lifecycle of multiple LLM models
type Manager struct {
	models     []ModelProvider // Ordered list of models from config file
	mu         sync.RWMutex
	config     *config.Config
	translator *i18n.I18n
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		models:     make([]ModelProvider, 0),
		config:     cfg,
		translator: i18n.New(cfg.GetLanguage()),
	}
}

func (m *Manager) RegisterModel(model ModelProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if model already exists
	for _, existing := range m.models {
		if existing.GetName() == model.GetName() {
			return
		}
	}

	// If this is the default model, insert at the beginning
	defaultModel := m.config.GetDefaultModel()
	if defaultModel != "" && model.GetName() == defaultModel {
		m.models = append([]ModelProvider{model}, m.models...)
		return
	}

	// Otherwise, append to the end
	m.models = append(m.models, model)
}

func (m *Manager) CallWithFallback(ctx context.Context, prompt string) (string, string, error) {
	return m.CallWithFallbackSystemPrompt(ctx, "", prompt)
}

func (m *Manager) CallWithFallbackSystemPrompt(ctx context.Context, systemPrompt string, userPrompt string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.models) == 0 {
		return "", "", fmt.Errorf("no LLM models configured")
	}

	for _, model := range m.models {
		// Check timeout context
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		default:
		}

		// Attempt to call
		var response string
		var err error

		// Start spinner before calling the model
		stopSpinner := ui.StartSpinner(m.translator.T("interactive.thinking"))

		// If model supports system prompt, use the version with system prompt
		if compatModel, ok := model.(*OpenAICompatibleProvider); ok && systemPrompt != "" {
			response, err = compatModel.CallWithSystemPrompt(ctx, systemPrompt, userPrompt)
		} else {
			response, err = model.Call(ctx, userPrompt)
		}

		// Stop spinner after the call completes
		if stopSpinner != nil {
			stopSpinner()
		}

		if err != nil {
			color.Red("Error: %v\n", err)
			continue
		}

		return response, model.GetName(), nil
	}

	return "", "", fmt.Errorf("all model calls failed")
}

func (m *Manager) GetStatus() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]map[string]interface{})

	for _, model := range m.models {
		status[model.GetName()] = map[string]interface{}{
			"name": model.GetName(),
		}
	}

	return status
}

func (m *Manager) PrintStatus() {
	status := m.GetStatus()

	fmt.Println("\n[" + m.translator.T("llm.status_title") + "]")

	for modelName := range status {
		fmt.Printf("- %s\n", modelName)
	}
}
