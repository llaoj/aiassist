package llm

import (
	"context"
)

// ModelProvider defines the LLM model provider interface
type ModelProvider interface {
	// Call sends an API request
	Call(ctx context.Context, prompt string) (string, error)
	GetName() string
}
