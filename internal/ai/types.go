package ai

type MessagePart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type ChatMessage struct {
	Role    string        `json:"role"`
	Content []MessagePart `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type AgentStep struct {
	Observation string    `json:"observation"`
	Plan        string    `json:"plan"`
	Commands    []Command `json:"commands"`
}
