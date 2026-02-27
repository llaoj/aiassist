package llm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// OpenAICompatibleProvider is a universal provider for OpenAI-compatible APIs
// This can work with any LLM service that implements the OpenAI API standard
type OpenAICompatibleProvider struct {
	name       string
	baseURL    string
	apiKey     string
	modelName  string
	available  bool
	httpClient *http.Client
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
	// Create HTTP client with reasonable timeout settings
	// Timeout includes connection, request, and response time
	transport := &http.Transport{
		// Connection timeout
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // Connection establishment timeout
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// TLS handshake timeout
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // Secure by default
		},
		TLSHandshakeTimeout: 10 * time.Second,
		// Note: ResponseHeaderTimeout removed - let http.Client.Timeout handle overall timeout
		// AI APIs may take time to process large requests before sending response headers
		// Connection pool settings
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		// Proxy will be set via SetProxyFunc
	}

	client := &http.Client{
		Timeout:   120 * time.Second, // Total request timeout (increased for AI APIs)
		Transport: transport,
	}

	return &OpenAICompatibleProvider{
		name:       name,
		baseURL:    baseURL,
		apiKey:     apiKey,
		modelName:  modelName,
		available:  true,
		httpClient: client,
	}
}

// SetProxyFunc configures proxy function for the provider
// Use http.ProxyFromEnvironment for automatic environment-based proxy selection
// or http.ProxyURL for a fixed proxy URL
func (o *OpenAICompatibleProvider) SetProxyFunc(proxyFunc func(*http.Request) (*url.URL, error)) error {
	if proxyFunc == nil {
		return nil
	}

	transport, ok := o.httpClient.Transport.(*http.Transport)
	if !ok {
		return fmt.Errorf("transport is not *http.Transport")
	}

	transport.Proxy = proxyFunc
	return nil
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

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		// Check if it's a timeout error
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return "", fmt.Errorf("%s API call timeout: %w", o.name, err)
		}
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
