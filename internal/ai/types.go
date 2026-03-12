package ai

import "github.com/cvcvka5/scraipy/internal/bridge"

// MessagePart represents a structured piece of a chat message,
// supporting both text-based instructions and multi-modal input (images).
type MessagePart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

// ChatMessage represents a single turn in the conversation history.
type ChatMessage struct {
	Role    string        `json:"role"`
	Content []MessagePart `json:"content"`
}

// ChatRequest defines the payload structure for the AI provider.
type ChatRequest struct {
	Model          string            `json:"model"`
	Messages       []ChatMessage     `json:"messages"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

// ChatResponse parses the structured JSON output returned by the LLM.
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// AgentStep represents the schema the AI is strictly forced to follow.
// By using a dedicated struct, we ensure that the marshaling of the
// AI's internal reasoning always matches our expected JSON format.
type AgentStep struct {
	Observation string    `json:"observation"`
	Plan        string    `json:"plan"`
	Commands    []Command `json:"commands"`
}

// Command encapsulates an action triggered by the AI and its arguments.
type Command struct {
	Action    bridge.BrowserAction `json:"action"`
	Arguments []any                `json:"arguments"`
}
