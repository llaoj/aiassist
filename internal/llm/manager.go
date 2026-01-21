package llm

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/llaoj/aiassist/internal/config"
)

// Manager 管理多个 LLM 提供商的生命周期
type Manager struct {
	providers map[string]ModelProvider
	priority  map[string]int
	mu        sync.RWMutex
	config    *config.Config
}

// NewManager 创建新的 LLM 管理器
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		providers: make(map[string]ModelProvider),
		priority:  make(map[string]int),
		config:    cfg,
	}
}

// RegisterProvider 注册一个 LLM 提供商
func (m *Manager) RegisterProvider(name string, provider ModelProvider, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	m.priority[name] = priority
}

// CallWithFallback 调用主模型，失败时自动切换到备用模型
func (m *Manager) CallWithFallback(ctx context.Context, prompt string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 获取按优先级排序的可用提供商
	available := m.getAvailableProviders()
	if len(available) == 0 {
		return "", "", fmt.Errorf("没有可用的 LLM 提供商")
	}

	var lastErr error
	for _, providerName := range available {
		provider := m.providers[providerName]

		// 检查超时上下文
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		default:
		}

		// 尝试调用
		response, err := provider.Call(ctx, prompt)
		if err != nil {
			lastErr = err
			fmt.Printf("[警告] %s 调用失败，尝试下一个模型: %v\n", providerName, err)
			continue
		}

		return response, providerName, nil
	}

	return "", "", fmt.Errorf("所有模型调用失败: %w", lastErr)
}

// CallSpecific 调用指定模型
func (m *Manager) CallSpecific(ctx context.Context, modelName string, prompt string) (string, error) {
	m.mu.RLock()
	provider, exists := m.providers[modelName]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("模型 %s 不存在", modelName)
	}

	if !provider.IsAvailable() {
		return "", fmt.Errorf("模型 %s 额度已用尽或不可用", modelName)
	}

	return provider.Call(ctx, prompt)
}

// GetAvailableProviders 获取可用的提供商列表（按优先级排序）
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

	// 按优先级排序（降序）
	sort.Slice(available, func(i, j int) bool {
		return available[i].priority > available[j].priority
	})

	result := make([]string, len(available))
	for i, p := range available {
		result[i] = p.name
	}

	return result
}

// GetStatus 获取所有提供商的状态信息
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

// ResetDailyQuota 重置每日调用配额（应在每天的指定时间调用）
func (m *Manager) ResetDailyQuota() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, modelCfg := range m.config.Models {
		if !modelCfg.Enabled {
			continue
		}

		if _, exists := m.providers[name]; exists {
			modelCfg.CurrentCalls = 0
			modelCfg.LastResetTime = time.Now().Unix()
		}
	}

	return m.config.Save()
}

// PrintStatus 打印当前模型状态到终端
func (m *Manager) PrintStatus() {
	status := m.GetStatus()

	fmt.Println("\n当前模型状态:")
	fmt.Println("─────────────────────────────────────────")

	for modelName, info := range status {
		available := info["available"].(bool)
		remainingCalls := info["remaining_calls"].(int)
		priority := info["priority"].(int)

		statusStr := "✓ 可用"
		if !available {
			statusStr = "✗ 不可用"
		}

		fmt.Printf("%s: %s | 剩余额度: %d | 优先级: %d\n",
			modelName, statusStr, remainingCalls, priority)
	}

	fmt.Println("─────────────────────────────────────────\n")
}
