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

type AnthropicProvider struct {
	model  models.Model
	apiKey string
}

func NewAnthropicProvider(model models.Model, apiKey string) *AnthropicProvider {
	return &AnthropicProvider{model: model, apiKey: apiKey}
}

func (p *AnthropicProvider) Model() models.Model {
	return p.model
}

type anthropicMessage struct {
	Role    string               `json:"role"`
	Content []anthropicContent   `json:"content"`
}

type anthropicContent struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Name  string `json:"name,omitempty"`
	ID    string `json:"id,omitempty"`
	Input any    `json:"input,omitempty"`
}

type anthropicRequest struct {
	Model       string              `json:"model"`
	MaxTokens   int64               `json:"max_tokens"`
	System      string              `json:"system,omitempty"`
	Messages    []anthropicMessage  `json:"messages"`
	Tools       []anthropicTool     `json:"tools,omitempty"`
	Stream      bool                `json:"stream"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	StopReason string         `json:"stop_reason"`
}

type anthropicStreamChunk struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta *struct {
		Text        string `json:"text,omitempty"`
		PartialJSON string `json:"partial_json,omitempty"`
		StopReason  string `json:"stop_reason,omitempty"`
	} `json:"delta,omitempty"`
	ContentBlock *struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input any    `json:"input,omitempty"`
		Text  string `json:"text,omitempty"`
	} `json:"content_block,omitempty"`
	Message *struct {
		ID string `json:"id"`
	} `json:"message,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func toAnthropicMessages(messages []Message) ([]anthropicMessage, string) {
	var result []anthropicMessage
	var system string
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		content := []anthropicContent{}
		if m.Content != "" {
			content = append(content, anthropicContent{Type: "text", Text: m.Content})
		}
		for _, tc := range m.ToolCalls {
			var input any
			json.Unmarshal([]byte(tc.Input), &input)
			content = append(content, anthropicContent{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: input,
			})
		}
		for _, tr := range m.ToolResults {
			content = append(content, anthropicContent{
				Type:  "tool_result",
				ID:    tr.ID,
				Text:  tr.Content,
			})
		}
		if len(content) == 0 {
			continue
		}
		result = append(result, anthropicMessage{Role: m.Role, Content: content})
	}
	return result, system
}

func toAnthropicTools(tools []ToolInfo) []anthropicTool {
	var result []anthropicTool
	for _, t := range tools {
		result = append(result, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: map[string]any{
				"type":       "object",
				"properties": t.Parameters,
				"required":   t.Required,
			},
		})
	}
	return result
}

func (p *AnthropicProvider) SendMessages(ctx context.Context, messages []Message, tools []ToolInfo) (*Message, error) {
	anMsg, system := toAnthropicMessages(messages)
	req := anthropicRequest{
		Model:     p.model.ID,
		MaxTokens: p.model.DefaultMaxTokens,
		System:    system,
		Messages:  anMsg,
		Tools:     toAnthropicTools(tools),
		Stream:    false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var anResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anResp); err != nil {
		return nil, err
	}

	msg := &Message{Role: "assistant"}
	for _, c := range anResp.Content {
		switch c.Type {
		case "text":
			msg.Content += c.Text
		case "tool_use":
			inputJSON, _ := json.Marshal(c.Input)
			msg.ToolCalls = append(msg.ToolCalls, ToolCall{
				ID:    c.ID,
				Name:  c.Name,
				Input: string(inputJSON),
			})
		}
	}
	return msg, nil
}

func (p *AnthropicProvider) StreamResponse(ctx context.Context, messages []Message, tools []ToolInfo) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent, 100)
	anMsg, system := toAnthropicMessages(messages)
	req := anthropicRequest{
		Model:     p.model.ID,
		MaxTokens: p.model.DefaultMaxTokens,
		System:    system,
		Messages:  anMsg,
		Tools:     toAnthropicTools(tools),
		Stream:    true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", p.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			ch <- ProviderEvent{Type: EventError, Err: fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(respBody))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var currentBlockType string
		var toolID, toolName string
		var toolInputBuf strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}
			var chunk anthropicStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Error != nil {
				ch <- ProviderEvent{Type: EventError, Err: fmt.Errorf("%s: %s", chunk.Error.Type, chunk.Error.Message)}
				return
			}

			switch chunk.Type {
			case "message_start":
			case "content_block_start":
				if chunk.ContentBlock != nil {
					currentBlockType = chunk.ContentBlock.Type
					if currentBlockType == "tool_use" {
						toolID = chunk.ContentBlock.ID
						toolName = chunk.ContentBlock.Name
						toolInputBuf.Reset()
						ch <- ProviderEvent{Type: EventToolUseStart, Name: toolName, ID: toolID}
					}
				}
			case "content_block_delta":
				if chunk.Delta != nil {
					if currentBlockType == "text" && chunk.Delta.Text != "" {
						ch <- ProviderEvent{Type: EventContentDelta, Content: chunk.Delta.Text}
					}
					if currentBlockType == "tool_use" && chunk.Delta.PartialJSON != "" {
						toolInputBuf.WriteString(chunk.Delta.PartialJSON)
						ch <- ProviderEvent{Type: EventToolUseDelta, Input: chunk.Delta.PartialJSON}
					}
				}
			case "content_block_stop":
				if currentBlockType == "tool_use" {
					ch <- ProviderEvent{Type: EventToolUseStop, Name: toolName, ID: toolID, Input: toolInputBuf.String()}
				}
				if currentBlockType == "text" {
					ch <- ProviderEvent{Type: EventContentStop}
				}
				currentBlockType = ""
			case "message_delta":
				if chunk.Delta != nil && chunk.Delta.StopReason == "end_turn" {
					ch <- ProviderEvent{Type: EventComplete}
				}
			case "message_stop":
			}
		}
	}()

	return ch, nil
}
