package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/n1xcode/n1x/internal/llm/models"
)

type OpenAIProvider struct {
	model   models.Model
	apiKey  string
	baseURL string
}

func NewOpenAIProvider(model models.Model, apiKey string, baseURL string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{model: model, apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *OpenAIProvider) Model() models.Model {
	return p.model
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	Index    int              `json:"index,omitempty"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function openAIFunction   `json:"function"`
}

type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIRequest struct {
	Model       string            `json:"model"`
	Messages    []openAIMessage   `json:"messages"`
	Tools       []openAITool      `json:"tools,omitempty"`
	Stream      bool              `json:"stream"`
	MaxTokens   int64             `json:"max_tokens,omitempty"`
}

type openAITool struct {
	Type     string           `json:"type"`
	Function openAIToolFunc   `json:"function"`
}

type openAIToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type openAIChoice struct {
	Delta struct {
		Role      string           `json:"role,omitempty"`
		Content   string           `json:"content,omitempty"`
		ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type openAIStreamChunk struct {
	Choices []openAIChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func toOpenAIMessages(messages []Message) []openAIMessage {
	var result []openAIMessage
	for _, m := range messages {
		switch m.Role {
		case "system":
			result = append(result, openAIMessage{Role: "system", Content: m.Content})
		case "user":
			result = append(result, openAIMessage{Role: "user", Content: m.Content})
		case "assistant":
			msg := openAIMessage{Role: "assistant", Content: m.Content}
			if len(m.ToolCalls) > 0 {
				var calls []openAIToolCall
				for _, tc := range m.ToolCalls {
					calls = append(calls, openAIToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: openAIFunction{
							Name:      tc.Name,
							Arguments: tc.Input,
						},
					})
				}
				msg.ToolCalls = calls
			}
			result = append(result, msg)
		case "tool":
			result = append(result, openAIMessage{
				Role:       "tool",
				Content:    m.Content,
				ToolCallID: m.ToolResults[0].ID,
			})
		}
	}
	return result
}

func toOpenAITools(tools []ToolInfo) []openAITool {
	var result []openAITool
	for _, t := range tools {
		result = append(result, openAITool{
			Type: "function",
			Function: openAIToolFunc{
				Name:        t.Name,
				Description: t.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": t.Parameters,
					"required":   t.Required,
				},
			},
		})
	}
	return result
}

func (p *OpenAIProvider) SendMessages(ctx context.Context, messages []Message, tools []ToolInfo) (*Message, error) {
	req := openAIRequest{
		Model:     p.model.ID,
		Messages:  toOpenAIMessages(messages),
		Tools:     toOpenAITools(tools),
		Stream:    false,
		MaxTokens: p.model.DefaultMaxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Role      string           `json:"role"`
				Content   string           `json:"content"`
				ToolCalls []openAIToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	msg := &Message{Role: "assistant", Content: result.Choices[0].Message.Content}
	for _, tc := range result.Choices[0].Message.ToolCalls {
		msg.ToolCalls = append(msg.ToolCalls, ToolCall{
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: tc.Function.Arguments,
		})
	}
	return msg, nil
}

func (p *OpenAIProvider) StreamResponse(ctx context.Context, messages []Message, tools []ToolInfo) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent, 100)
	req := openAIRequest{
		Model:     p.model.ID,
		Messages:  toOpenAIMessages(messages),
		Tools:     toOpenAITools(tools),
		Stream:    true,
		MaxTokens: p.model.DefaultMaxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			ch <- ProviderEvent{Type: EventError, Err: fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(respBody))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var toolCallAccumulator = make(map[int]struct {
			id   string
			name string
			buf  strings.Builder
		})

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- ProviderEvent{Type: EventComplete}
				continue
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Error != nil {
				ch <- ProviderEvent{Type: EventError, Err: fmt.Errorf("%s", chunk.Error.Message)}
				return
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					ch <- ProviderEvent{Type: EventContentDelta, Content: choice.Delta.Content}
				}
				for _, tc := range choice.Delta.ToolCalls {
					acc := toolCallAccumulator[tc.Index]
					if tc.ID != "" {
						acc.id = tc.ID
					}
					if tc.Function.Name != "" {
						acc.name = tc.Function.Name
					}
					if tc.Function.Arguments != "" {
						acc.buf.WriteString(tc.Function.Arguments)
					}
					if tc.Function.Name != "" || tc.ID != "" {
						ch <- ProviderEvent{Type: EventToolUseDelta, Name: tc.Function.Name, ID: tc.ID, Input: tc.Function.Arguments}
					}
					toolCallAccumulator[tc.Index] = acc
				}
				if choice.FinishReason == "tool_calls" {
					for _, acc := range toolCallAccumulator {
						ch <- ProviderEvent{Type: EventToolUseStop, Name: acc.name, ID: acc.id, Input: acc.buf.String()}
					}
					toolCallAccumulator = make(map[int]struct {
						id   string
						name string
						buf  strings.Builder
					})
				}
			}
		}
	}()

	return ch, nil
}
