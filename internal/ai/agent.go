package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	MaxHistoryItems = 6 // Reduced even further to aggressively fight 400 errors
	MaxRetries      = 5
	InitialBackoff  = 2 * time.Second
)

type Agent struct {
	mu sync.RWMutex

	Key          string        `json:"api_key"`
	Model        string        `json:"model"`
	SystemPrompt string        `json:"system_prompt"`
	Messages     []ChatMessage `json:"messages"`
}

func NewAgent(key, model, systemPrompt string) *Agent {
	return &Agent{
		Key:          key,
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages:     make([]ChatMessage, 0),
	}
}

func (a *Agent) AddMessage(role string, parts ...MessagePart) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Messages = append(a.Messages, ChatMessage{
		Role:    role,
		Content: parts,
	})
}

func (a *Agent) Send(role string, parts ...MessagePart) (*ChatResponse, error) {
	a.AddMessage(role, parts...)

	a.mu.RLock()
	startIdx := 0
	if len(a.Messages) > MaxHistoryItems {
		startIdx = len(a.Messages) - MaxHistoryItems
	}

	payloadMessages := make([]ChatMessage, 0, (len(a.Messages)-startIdx)+1)
	payloadMessages = append(payloadMessages, ChatMessage{
		Role:    "system",
		Content: []MessagePart{{Type: "text", Text: a.SystemPrompt}},
	})
	payloadMessages = append(payloadMessages, a.Messages[startIdx:]...)
	a.mu.RUnlock()

	reqBody := ChatRequest{
		Model:          a.Model,
		Messages:       payloadMessages,
		ResponseFormat: map[string]string{"type": "json_object"},
	}

	resp, err := a.executeWithRetry(context.Background(), reqBody)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) > 0 {
		aiMsg := resp.Choices[0].Message
		a.AddMessage("assistant", MessagePart{Type: "text", Text: aiMsg.Content})
	}

	return resp, nil
}

func (a *Agent) executeWithRetry(ctx context.Context, body ChatRequest) (*ChatResponse, error) {
	var lastErr error
	backoff := InitialBackoff

	for i := 0; i < MaxRetries; i++ {
		httpResp, err := a.doPost(ctx, body)

		if err == nil && httpResp.StatusCode == http.StatusOK {
			defer httpResp.Body.Close()
			var result ChatResponse
			if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
				return nil, fmt.Errorf("decode error: %w", err)
			}
			return &result, nil
		}

		// --- VERBOSE DEBUGGING LOGIC ---
		if httpResp != nil {
			bodyBytes, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()

			// Print the verbose error from the API provider
			fmt.Printf("\n[DEBUG] API Status: %d\n", httpResp.StatusCode)
			fmt.Printf("[DEBUG] API Error Response: %s\n", string(bodyBytes))

			// Check request size
			reqJson, _ := json.Marshal(body)
			fmt.Printf("[DEBUG] Payload Size: %d bytes\n", len(reqJson))

			lastErr = fmt.Errorf("status %d: %s", httpResp.StatusCode, string(bodyBytes))
		} else if err != nil {
			lastErr = err
		}

		if !isRetryable(httpResp, err) {
			break
		}

		select {
		case <-time.After(backoff):
			backoff *= 2
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("agent failed: %w", lastErr)
}

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

	return http.DefaultClient.Do(req)
}

func (a *Agent) SendText(role, txt string) (*ChatResponse, error) {
	return a.Send(role, MessagePart{Type: "text", Text: txt})
}

func getStatus(r *http.Response) int {
	if r != nil {
		return r.StatusCode
	}
	return 0
}

func isRetryable(r *http.Response, err error) bool {
	if err != nil {
		return true
	}
	s := r.StatusCode
	return s == 429 || s >= 500 || s == 408
}
