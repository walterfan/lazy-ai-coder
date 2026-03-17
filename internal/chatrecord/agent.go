package chatrecord

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// Agent defines the interface for the learning record AI agent
// that classifies user input and generates type-appropriate responses.
type Agent interface {
	// Process takes user input, classifies it, and generates a response.
	// It returns the classified input type and the generated response payload.
	Process(ctx context.Context, input string) (*ProcessResult, error)
	// ProcessWithHistory does the same with conversation history for multi-turn context.
	// history is prior user/assistant messages (e.g. from session). May be nil or empty.
	ProcessWithHistory(ctx context.Context, input string, history []*schema.Message) (*ProcessResult, error)
}

// ProcessResult contains the result of processing user input
type ProcessResult struct {
	InputType       string                      `json:"input_type"`
	ResponsePayload *models.ResponsePayloadData `json:"response_payload"`
}

// AgentConfig holds configuration for the learning record agent
type AgentConfig struct {
	// LLM settings
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// DefaultAgentConfig returns a default configuration
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2048,
	}
}
