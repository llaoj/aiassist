package llm

import (
	"context"
)

// Model defines the LLM model interface
type Model interface {
	// Call sends an API request
	Call(ctx context.Context, prompt string) (string, error)
	GetName() string
}
