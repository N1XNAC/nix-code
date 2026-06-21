package provider

import (
	"context"
	"errors"

	"github.com/n1xcode/n1x/internal/llm/models"
)

type EventType string

const (
	EventContentStart  EventType = "content_start"
	EventContentDelta  EventType = "content_delta"
	EventThinkingDelta EventType = "thinking_delta"
	EventToolUseStart  EventType = "tool_use_start"
	EventToolUseDelta  EventType = "tool_use_delta"
	EventToolUseStop   EventType = "tool_use_stop"
	EventContentStop   EventType = "content_stop"
	EventComplete      EventType = "complete"
	EventError         EventType = "error"
)

type ProviderEvent struct {
	Type    EventType
	Content string
	Name    string
	Input   string
	ID      string
	Err     error
}

type Message struct {
	Role    string       `json:"role"`
	Content string       `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`
}

type ToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Input string `json:"input"`
}

type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

type ToolInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
	Required    []string       `json:"required"`
}

type Provider interface {
	SendMessages(ctx context.Context, messages []Message, tools []ToolInfo) (*Message, error)
	StreamResponse(ctx context.Context, messages []Message, tools []ToolInfo) (<-chan ProviderEvent, error)
	Model() models.Model
}

var openAICompatibleProviders = map[models.ModelProvider]bool{
	models.ProviderOpenAI:     true,
	models.ProviderOpenRouter: true,
	models.ProviderGroq:      true,
	models.ProviderNvidiaNIM: true,
	models.ProviderKimi:      true,
	models.ProviderGLM:       true,
	models.ProviderDeepSeek:  true,
	models.ProviderMistral:   true,
	models.ProviderTogether:  true,
	models.ProviderFireworks: true,
	models.ProviderPerplexity: true,
	models.ProviderAnyscale:  true,
	models.ProviderXAI:       true,
	models.ProviderCohere:    true,
	models.ProviderVoyage:    true,
	models.ProviderAI21:      true,
}

func NewProvider(model models.Model, apiKey string, baseURL string) (Provider, error) {
	switch model.Provider {
	case models.ProviderAnthropic:
		return NewAnthropicProvider(model, apiKey), nil
	case models.ProviderOpenAI:
		return NewOpenAIProvider(model, apiKey, baseURL), nil
	case models.ProviderGemini:
		return NewGeminiProvider(model, apiKey), nil
	default:
		if openAICompatibleProviders[model.Provider] {
			if baseURL == "" {
				baseURL = model.Provider.DefaultBaseURL()
			}
			return NewOpenAIProvider(model, apiKey, baseURL), nil
		}
		return nil, errors.New("unsupported provider: " + string(model.Provider))
	}
}
