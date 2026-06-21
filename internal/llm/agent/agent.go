package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/n1xcode/n1x/internal/config"
	"github.com/n1xcode/n1x/internal/llm/models"
	"github.com/n1xcode/n1x/internal/llm/provider"
	"github.com/n1xcode/n1x/internal/llm/tools"
	"github.com/n1xcode/n1x/internal/permission"
	"github.com/n1xcode/n1x/internal/pubsub"
)

type Mode string

const (
	ModeThink Mode = "think"
	ModeCode  Mode = "code"
)

type Agent struct {
	mu          sync.RWMutex
	mode        Mode
	provider    provider.Provider
	model       models.Model
	tools       []tools.Tool
	permissions *permission.PermissionService
	bus         *pubsub.Bus
	systemPrompt string
}

func New(cfg *config.Config, bus *pubsub.Bus, perms *permission.PermissionService) (*Agent, error) {
	providerName, providerCfg, ok := cfg.GetActiveProvider()
	if !ok {
		return nil, fmt.Errorf("no API key configured. Run 'nix config' to set up your API keys")
	}

	model, ok := models.FindModel(providerCfg.Model)
	if !ok {
		model, ok = models.DefaultModelForProvider(models.ModelProvider(providerName))
		if !ok {
			return nil, fmt.Errorf("no default model found for provider: %s", providerName)
		}
	}

	p, err := provider.NewProvider(model, providerCfg.APIKey, providerCfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	builtinTools := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(),
		tools.NewEditTool(),
		tools.NewBashTool(),
		tools.NewGrepTool(),
		tools.NewGlobTool(),
		tools.NewTodoWriteTool(),
	}

	systemPrompt := `You are N1X Code, a terminal-based AI coding agent.

You have access to tools that let you read, write, and edit files, run commands, search code, and manage tasks.

Rules:
1. Always think about what the user needs before acting
2. Use todowrite to track progress on complex multi-step tasks
3. Prefer edit over write for making targeted changes to existing files
4. Read files before editing them to understand the full context
5. Use grep to find relevant code before making changes
6. For bash commands, prefer safe read-only commands first
7. Ask clarifying questions when requirements are ambiguous
8. Show your work and explain what you're doing`

	return &Agent{
		mode:         ModeCode,
		provider:     p,
		model:        model,
		tools:        builtinTools,
		permissions:  perms,
		bus:          bus,
		systemPrompt: systemPrompt,
	}, nil
}

func (a *Agent) Mode() Mode {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.mode
}

func (a *Agent) SetMode(m Mode) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.mode = m
	a.permissions.SetAgentPermissions(string(m))
	if a.bus != nil {
		a.bus.Publish(pubsub.Event{
			Type: pubsub.EventModeChanged,
			Data: map[string]any{"mode": string(m)},
		})
	}
}

func (a *Agent) Tools() []tools.Tool {
	return a.tools
}

func (a *Agent) Provider() provider.Provider {
	return a.provider
}

func (a *Agent) Run(ctx context.Context, userPrompt string, sessionID string) (string, error) {
	messages := []provider.Message{
		{Role: "system", Content: a.systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return a.loop(ctx, messages, sessionID)
}

func (a *Agent) loop(ctx context.Context, messages []provider.Message, sessionID string) (string, error) {
	maxIterations := 25

	for i := 0; i < maxIterations; i++ {
		response, err := a.provider.SendMessages(ctx, messages, a.getToolInfos())
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		messages = append(messages, *response)

		if len(response.ToolCalls) == 0 {
			return response.Content, nil
		}

		for _, tc := range response.ToolCalls {
			allowed, err := a.permissions.Check(tc.Name, sessionID)
			if !allowed {
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				} else {
					errMsg = fmt.Sprintf("tool '%s' is not allowed", tc.Name)
				}
				messages = append(messages, provider.Message{
					Role: "tool",
					ToolResults: []provider.ToolResult{{
						ID:      tc.ID,
						Content: errMsg,
						Error:   errMsg,
					}},
				})
				continue
			}

			if a.bus != nil {
				a.bus.Publish(pubsub.Event{
					Type:    pubsub.EventToolStarted,
					Session: sessionID,
					Data:    map[string]any{"tool": tc.Name, "input": tc.Input},
				})
			}

			tool := tools.FindTool(a.tools, tc.Name)
			if tool == nil {
				errMsg := fmt.Sprintf("unknown tool: %s", tc.Name)
				messages = append(messages, provider.Message{
					Role: "tool",
					ToolResults: []provider.ToolResult{{
						ID:      tc.ID,
						Content: errMsg,
						Error:   errMsg,
					}},
				})
				continue
			}

			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Input), &args); err != nil {
				errMsg := fmt.Sprintf("invalid tool arguments: %s", err)
				messages = append(messages, provider.Message{
					Role: "tool",
					ToolResults: []provider.ToolResult{{
						ID:      tc.ID,
						Content: errMsg,
						Error:   errMsg,
					}},
				})
				continue
			}

			result, err := tool.Execute(ctx, args, sessionID)
			if err != nil {
				messages = append(messages, provider.Message{
					Role: "tool",
					ToolResults: []provider.ToolResult{{
						ID:      tc.ID,
						Content: err.Error(),
						Error:   err.Error(),
					}},
				})
			} else {
				messages = append(messages, provider.Message{
					Role: "tool",
					ToolResults: []provider.ToolResult{{
						ID:      tc.ID,
						Content: result,
					}},
				})
			}

			if a.bus != nil {
				a.bus.Publish(pubsub.Event{
					Type:    pubsub.EventToolCompleted,
					Session: sessionID,
					Data:    map[string]any{"tool": tc.Name},
				})
			}
		}
	}

	return "", fmt.Errorf("agent reached maximum iterations (%d)", maxIterations)
}

func (a *Agent) getToolInfos() []provider.ToolInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var allowedTools []tools.Tool
	for _, t := range a.tools {
		level := a.permissions.GetRule(t.Name())
		if level != permission.Deny {
			allowedTools = append(allowedTools, t)
		}
	}

	return tools.ToToolInfos(allowedTools)
}

func (a *Agent) StreamRun(ctx context.Context, userPrompt string, sessionID string, eventCh chan<- provider.ProviderEvent) (string, error) {
	messages := []provider.Message{
		{Role: "system", Content: a.systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return a.streamLoop(ctx, messages, sessionID, eventCh)
}

func (a *Agent) streamLoop(ctx context.Context, messages []provider.Message, sessionID string, eventCh chan<- provider.ProviderEvent) (string, error) {
	maxIterations := 25
	var fullContent strings.Builder

	for i := 0; i < maxIterations; i++ {
		streamCh, err := a.provider.StreamResponse(ctx, messages, a.getToolInfos())
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		var response provider.Message
		var toolCallBuf strings.Builder
		var currentTool provider.ToolCall

		for event := range streamCh {
			switch event.Type {
			case provider.EventContentDelta:
				response.Content += event.Content
				fullContent.WriteString(event.Content)
				if eventCh != nil {
					eventCh <- event
				}

			case provider.EventToolUseStart:
				currentTool = provider.ToolCall{
					ID:    event.ID,
					Name:  event.Name,
					Input: event.Input,
				}
				toolCallBuf.Reset()
				if eventCh != nil {
					eventCh <- event
				}

			case provider.EventToolUseDelta:
				toolCallBuf.WriteString(event.Input)
				if eventCh != nil {
					eventCh <- event
				}

			case provider.EventToolUseStop:
				currentTool.Input = toolCallBuf.String()
				response.ToolCalls = append(response.ToolCalls, currentTool)

			case provider.EventComplete:
				response.Role = "assistant"
				messages = append(messages, response)

				if len(response.ToolCalls) == 0 {
					if eventCh != nil {
						close(eventCh)
					}
					return fullContent.String(), nil
				}

				for _, tc := range response.ToolCalls {
					if eventCh != nil {
						eventCh <- provider.ProviderEvent{
							Type: provider.EventToolUseStart,
							Name: tc.Name,
							ID:   tc.ID,
							Input: tc.Input,
						}
					}

					allowed, err := a.permissions.Check(tc.Name, sessionID)
					if !allowed {
						errMsg := fmt.Sprintf("tool '%s' is not allowed in %s mode", tc.Name, a.Mode())
						messages = append(messages, provider.Message{
							Role: "tool",
							ToolResults: []provider.ToolResult{{
								ID:      tc.ID,
								Content: errMsg,
								Error:   errMsg,
							}},
						})
						continue
					}

					tool := tools.FindTool(a.tools, tc.Name)
					if tool == nil {
						errMsg := fmt.Sprintf("unknown tool: %s", tc.Name)
						messages = append(messages, provider.Message{
							Role: "tool",
							ToolResults: []provider.ToolResult{{
								ID:      tc.ID,
								Content: errMsg,
								Error:   errMsg,
							}},
						})
						continue
					}

					var args map[string]any
					if err := json.Unmarshal([]byte(tc.Input), &args); err != nil {
						errMsg := fmt.Sprintf("invalid arguments: %s", err)
						messages = append(messages, provider.Message{
							Role: "tool",
							ToolResults: []provider.ToolResult{{
								ID:      tc.ID,
								Content: errMsg,
								Error:   errMsg,
							}},
						})
						continue
					}

					result, err := tool.Execute(ctx, args, sessionID)
					if err != nil {
						messages = append(messages, provider.Message{
							Role: "tool",
							ToolResults: []provider.ToolResult{{
								ID:      tc.ID,
								Content: err.Error(),
								Error:   err.Error(),
							}},
						})
					} else {
						messages = append(messages, provider.Message{
							Role: "tool",
							ToolResults: []provider.ToolResult{{
								ID:      tc.ID,
								Content: result,
							}},
						})
					}
				}

				response = provider.Message{}
				toolCallBuf.Reset()
			}
		}
	}

	return "", fmt.Errorf("agent reached maximum iterations (%d)", maxIterations)
}
