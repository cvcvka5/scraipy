package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Agent handles the stateful conversation with the LLM.
type Agent struct {
	mu sync.Mutex

	Key          string        `json:"api_key"`
	Model        string        `json:"model"`
	SystemPrompt string        `json:"system_prompt"`
	Messages     []ChatMessage `json:"messages"`
}

// NewAgent creates a fresh agent instance with empty history.
func NewAgent(key string, model string, systemPrompt string) *Agent {
	return &Agent{
		Key:          key,
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages:     []ChatMessage{},
	}
}

// AddMessage appends a message to history without triggering an API call.
// This is used for adding browser observations or manual error reports.
func (a *Agent) AddMessage(role string, parts ...MessagePart) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Messages = append(a.Messages, ChatMessage{
		Role:    role,
		Content: parts,
	})
}

// Send manages the full conversation lifecycle:
// 1. Appends the new user/tool message to history via AddMessage.
// 2. Sends the full context (System Prompt + History) to the AI.
// 3. Automatically appends the resulting AI's response to history as an 'assistant'.
func (a *Agent) Send(role string, parts ...MessagePart) (*ChatResponse, error) {
	// Add the incoming message to internal history first
	a.AddMessage(role, parts...)

	// Create a safe copy of messages including the system prompt for the request
	a.mu.Lock()
	msgCopy := make([]ChatMessage, 0, len(a.Messages)+1)
	msgCopy = append(msgCopy, ChatMessage{
		Role:    "system",
		Content: []MessagePart{{Type: "text", Text: a.SystemPrompt}},
	})
	msgCopy = append(msgCopy, a.Messages...)
	a.mu.Unlock()

	// Execute request with retry logic
	resp, err := a.executeWithRetry(context.Background(), ChatRequest{
		Model:    a.Model,
		Messages: msgCopy,
	})
	if err != nil {
		return nil, err
	}

	// Capture and save the Assistant's response to maintain conversation state
	if resp != nil && len(resp.Choices) > 0 {
		aiMsg := resp.Choices[0].Message
		a.AddMessage("assistant", MessagePart{Type: "text", Text: aiMsg.Content})
	}

	return resp, nil
}

// executeWithRetry handles transient API errors with exponential backoff.
func (a *Agent) executeWithRetry(ctx context.Context, body ChatRequest) (*ChatResponse, error) {
	var lastErr error
	backoff := 2 * time.Second

	for i := 0; i < 5; i++ {
		resp, err := a.doPost(ctx, body)

		// Success case
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var result ChatResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return nil, fmt.Errorf("json decode error: %w", err)
			}
			return &result, nil
		}

		// Error tracking
		lastErr = fmt.Errorf("status %d: %v", getStatus(resp), err)

		// Decide if we should retry
		if !isRetryable(resp, err) {
			if resp != nil {
				resp.Body.Close()
			}
			break
		}

		// Close response body before retrying
		if resp != nil {
			resp.Body.Close()
		}

		// Wait with backoff or context cancellation
		select {
		case <-time.After(backoff):
			backoff *= 2
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("request failed after 5 attempts: %w", lastErr)
}

// doPost performs the actual HTTP POST to the API endpoint.
func (a *Agent) doPost(ctx context.Context, body ChatRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Key)
	req.Header.Set("X-Title", "Scraipy Agent")

	return http.DefaultClient.Do(req)
}

// SendText is a convenience wrapper for plain text messages.
func (a *Agent) SendText(role, txt string) (*ChatResponse, error) {
	return a.Send(role, MessagePart{Type: "text", Text: txt})
}

// SendImage is a convenience wrapper for messages containing an image URL.
func (a *Agent) SendImage(role, txt, url string) (*ChatResponse, error) {
	return a.Send(role,
		MessagePart{Type: "text", Text: txt},
		MessagePart{Type: "image_url", ImageURL: &struct {
			URL string `json:"url"`
		}{URL: url}},
	)
}

// getStatus safely extracts the status code from a response.
func getStatus(r *http.Response) int {
	if r != nil {
		return r.StatusCode
	}
	return 0
}

// isRetryable determines if an error code warrants a retry attempt.
func isRetryable(r *http.Response, err error) bool {
	if err != nil {
		return true // Always retry network errors
	}
	s := r.StatusCode
	// Retry on Rate Limits (429), Server Errors (5xx), and Request Timeouts (408).
	return s == 429 || s >= 500 || s == 408
}
