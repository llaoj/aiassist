package llm

import (
	"context"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelProvider defines the LLM model provider interface
type ModelProvider interface {
	// Call sends an API request
	Call(ctx context.Context, prompt string) (string, error)
	GetName() string
	IsAvailable() bool
	GetRemainingCalls() int
}
