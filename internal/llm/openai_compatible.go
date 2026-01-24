package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAICompatibleProvider is a universal provider for OpenAI-compatible APIs
// This can work with any LLM service that implements the OpenAI API standard
type OpenAICompatibleProvider struct {
	name      string
	baseURL   string
	apiKey    string
	modelName string
	available bool
}

// Request and Response structures for OpenAI API
type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type choice struct {
	Message chatMessage `json:"message"`
}

// NewOpenAICompatibleProvider creates a new OpenAI-compatible provider
func NewOpenAICompatibleProvider(name, baseURL, apiKey, modelName string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		name:      name,
		baseURL:   baseURL,
		apiKey:    apiKey,
		modelName: modelName,
		available: true,
	}
}

func (o *OpenAICompatibleProvider) GetName() string {
	return o.name
}

func (o *OpenAICompatibleProvider) IsAvailable() bool {
	return o.available
}

func (o *OpenAICompatibleProvider) GetRemainingCalls() int {
	return 0 // No longer tracking quota
}

func (o *OpenAICompatibleProvider) Call(ctx context.Context, prompt string) (string, error) {
	return o.CallWithSystemPrompt(ctx, "", prompt)
}

// CallWithSystemPrompt calls the LLM API with a system prompt
func (o *OpenAICompatibleProvider) CallWithSystemPrompt(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	if !o.available {
		return "", fmt.Errorf("%s is unavailable (quota exhausted or billing issue)", o.name)
	}

	// Prepare messages
	messages := []chatMessage{}

	// Add system message if provided
	if systemPrompt != "" {
		messages = append(messages, chatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add user message
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: userPrompt,
	})

	// Prepare request
	req := chatCompletionRequest{
		Model:    o.modelName,
		Messages: messages,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%s API call failed: %w", o.name, err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode == 429 {
		// Rate limit or quota exceeded - mark as unavailable
		o.available = false
		return "", fmt.Errorf("%s: quota exceeded or rate limited (HTTP 429)", o.name)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var respData chatCompletionResponse
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if respData.Error != nil {
		return "", fmt.Errorf("API error from %s: %s", o.name, respData.Error.Message)
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("no response from %s", o.name)
	}

	return respData.Choices[0].Message.Content, nil
}
