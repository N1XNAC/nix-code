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

type GeminiProvider struct {
	model  models.Model
	apiKey string
}

func NewGeminiProvider(model models.Model, apiKey string) *GeminiProvider {
	return &GeminiProvider{model: model, apiKey: apiKey}
}

func (p *GeminiProvider) Model() models.Model {
	return p.model
}

type geminiContent struct {
	Role  string        `json:"role,omitempty"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text        string           `json:"text,omitempty"`
	FunctionCall *geminiFunction `json:"functionCall,omitempty"`
	FunctionResp *geminiFuncResp `json:"functionResponse,omitempty"`
}

type geminiFunction struct {
	Name  string `json:"name"`
	Args  any    `json:"args"`
}

type geminiFuncResp struct {
	Name     string `json:"name"`
	Response struct {
		Content any `json:"content"`
	} `json:"response"`
}

type geminiRequest struct {
	Contents         []geminiContent   `json:"contents"`
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	Tools            []geminiTool      `json:"tools,omitempty"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFuncDecl `json:"functionDeclarations"`
}

type geminiFuncDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

type geminiStreamChunk struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

func toGeminiMessages(messages []Message) ([]geminiContent, *geminiContent) {
	var contents []geminiContent
	var system *geminiContent

	for _, m := range messages {
		if m.Role == "system" {
			system = &geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			}
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		parts := []geminiPart{}
		if m.Content != "" {
			parts = append(parts, geminiPart{Text: m.Content})
		}
		for _, tc := range m.ToolCalls {
			var args any
			json.Unmarshal([]byte(tc.Input), &args)
			parts = append(parts, geminiPart{
				FunctionCall: &geminiFunction{Name: tc.Name, Args: args},
			})
		}
		for _, tr := range m.ToolResults {
			var respContent any = tr.Content
			if tr.Error != "" {
				respContent = map[string]string{"error": tr.Error}
			}
			parts = append(parts, geminiPart{
				FunctionResp: &geminiFuncResp{
					Name: tr.ID,
					Response: struct {
						Content any `json:"content"`
					}{Content: respContent},
				},
			})
		}
		contents = append(contents, geminiContent{Role: role, Parts: parts})
	}
	return contents, system
}

func toGeminiTools(tools []ToolInfo) []geminiTool {
	var decls []geminiFuncDecl
	for _, t := range tools {
		decls = append(decls, geminiFuncDecl{
			Name:        t.Name,
			Description: t.Description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": t.Parameters,
				"required":   t.Required,
			},
		})
	}
	return []geminiTool{{FunctionDeclarations: decls}}
}

func (p *GeminiProvider) SendMessages(ctx context.Context, messages []Message, tools []ToolInfo) (*Message, error) {
	contents, system := toGeminiMessages(messages)
	req := geminiRequest{Contents: contents, SystemInstruction: system}
	if len(tools) > 0 {
		req.Tools = toGeminiTools(tools)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model.ID, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, err
	}

	if len(gResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned")
	}

	msg := &Message{Role: "assistant"}
	for _, part := range gResp.Candidates[0].Content.Parts {
		if part.Text != "" {
			msg.Content += part.Text
		}
		if part.FunctionCall != nil {
			args, _ := json.Marshal(part.FunctionCall.Args)
			msg.ToolCalls = append(msg.ToolCalls, ToolCall{
				Name:  part.FunctionCall.Name,
				Input: string(args),
			})
		}
	}
	return msg, nil
}

func (p *GeminiProvider) StreamResponse(ctx context.Context, messages []Message, tools []ToolInfo) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent, 100)

	contents, system := toGeminiMessages(messages)
	req := geminiRequest{Contents: contents, SystemInstruction: system}
	if len(tools) > 0 {
		req.Tools = toGeminiTools(tools)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.model.ID, p.apiKey)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			ch <- ProviderEvent{Type: EventError, Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			ch <- ProviderEvent{Type: EventError, Err: fmt.Errorf("Gemini API error (%d): %s", resp.StatusCode, string(respBody))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
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

			var chunk geminiStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			for _, c := range chunk.Candidates {
				for _, part := range c.Content.Parts {
					if part.Text != "" {
						ch <- ProviderEvent{Type: EventContentDelta, Content: part.Text}
					}
					if part.FunctionCall != nil {
						args, _ := json.Marshal(part.FunctionCall.Args)
						ch <- ProviderEvent{Type: EventToolUseStart, Name: part.FunctionCall.Name, Input: string(args)}
					}
				}
			}
		}
	}()

	return ch, nil
}
