package llm

import (
	"context"
	"fmt"
)

// ModelProvider 定义 LLM 模型提供商接口
type ModelProvider interface {
	// Call 发起 API 调用
	Call(ctx context.Context, prompt string) (string, error)
	// GetName 获取模型名称
	GetName() string
	// IsAvailable 检查模型是否可用
	IsAvailable() bool
	// GetRemainingCalls 获取剩余调用次数
	GetRemainingCalls() int
}

// ChatGPTProvider OpenAI ChatGPT 提供商
type ChatGPTProvider struct {
	name           string
	apiKey         string
	proxy          string
	remainingCalls int
}

// NewChatGPTProvider 创建新的 ChatGPT 提供商
func NewChatGPTProvider(apiKey, proxy string) *ChatGPTProvider {
	return &ChatGPTProvider{
		name:           "ChatGPT",
		apiKey:         apiKey,
		proxy:          proxy,
		remainingCalls: 1000,
	}
}

func (c *ChatGPTProvider) GetName() string {
	return c.name
}

func (c *ChatGPTProvider) IsAvailable() bool {
	return c.remainingCalls > 0
}

func (c *ChatGPTProvider) GetRemainingCalls() int {
	return c.remainingCalls
}

func (c *ChatGPTProvider) Call(ctx context.Context, prompt string) (string, error) {
	if c.remainingCalls <= 0 {
		return "", fmt.Errorf("ChatGPT 额度已用尽")
	}

	// 调用 ChatGPT API（示例）
	// 实际实现应使用代理并发起 HTTP 请求
	response := fmt.Sprintf("ChatGPT 回复: %s", prompt)
	c.remainingCalls--
	return response, nil
}

// QianWenProvider 通义千问提供商
type QianWenProvider struct {
	name           string
	apiKey         string
	remainingCalls int
}

// NewQianWenProvider 创建新的通义千问提供商
func NewQianWenProvider(apiKey string) *QianWenProvider {
	return &QianWenProvider{
		name:           "通义千问",
		apiKey:         apiKey,
		remainingCalls: 1000,
	}
}

func (q *QianWenProvider) GetName() string {
	return q.name
}

func (q *QianWenProvider) IsAvailable() bool {
	return q.remainingCalls > 0
}

func (q *QianWenProvider) GetRemainingCalls() int {
	return q.remainingCalls
}

func (q *QianWenProvider) Call(ctx context.Context, prompt string) (string, error) {
	if q.remainingCalls <= 0 {
		return "", fmt.Errorf("通义千问额度已用尽")
	}

	// 调用通义千问 API（示例）
	// 实际实现应使用官方 SDK 或 HTTP 客户端
	response := fmt.Sprintf("通义千问回复: %s", prompt)
	q.remainingCalls--
	return response, nil
}

// DeepSeekProvider DeepSeek 代码模型提供商
type DeepSeekProvider struct {
	name           string
	apiKey         string
	remainingCalls int
}

// NewDeepSeekProvider 创建新的 DeepSeek 提供商
func NewDeepSeekProvider(apiKey string) *DeepSeekProvider {
	return &DeepSeekProvider{
		name:           "DeepSeek",
		apiKey:         apiKey,
		remainingCalls: 1000,
	}
}

func (d *DeepSeekProvider) GetName() string {
	return d.name
}

func (d *DeepSeekProvider) IsAvailable() bool {
	return d.remainingCalls > 0
}

func (d *DeepSeekProvider) GetRemainingCalls() int {
	return d.remainingCalls
}

func (d *DeepSeekProvider) Call(ctx context.Context, prompt string) (string, error) {
	if d.remainingCalls <= 0 {
		return "", fmt.Errorf("DeepSeek 额度已用尽")
	}

	// 调用 DeepSeek API（示例）
	response := fmt.Sprintf("DeepSeek 回复: %s", prompt)
	d.remainingCalls--
	return response, nil
}
