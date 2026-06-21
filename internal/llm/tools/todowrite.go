package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type TodoItem struct {
	Content  string `json:"content"`
	Status   string `json:"status"`
	Priority string `json:"priority,omitempty"`
}

type TodoStore struct {
	mu    sync.RWMutex
	todos map[string][]TodoItem
}

var GlobalTodoStore = &TodoStore{todos: make(map[string][]TodoItem)}

type TodoWriteTool struct{}

func NewTodoWriteTool() *TodoWriteTool {
	return &TodoWriteTool{}
}

func (t *TodoWriteTool) Name() string {
	return "todowrite"
}

func (t *TodoWriteTool) Description() string {
	return "Create and update task lists to track progress during complex operations."
}

func (t *TodoWriteTool) Parameters() map[string]any {
	return map[string]any{
		"todos": map[string]any{
			"type":        "array",
			"description": "List of todo items",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{
						"type":        "string",
						"description": "Description of the task",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Status: pending, in_progress, completed, cancelled",
						"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority: high, medium, low",
						"enum":        []string{"high", "medium", "low"},
					},
				},
				"required": []string{"content", "status"},
			},
		},
	}
}

func (t *TodoWriteTool) Required() []string {
	return []string{"todos"}
}

func (t *TodoWriteTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	todosRaw, ok := args["todos"].([]any)
	if !ok {
		return "", fmt.Errorf("todos must be an array")
	}

	var items []TodoItem
	for _, raw := range todosRaw {
		itemMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		item := TodoItem{
			Content:  getString(itemMap, "content"),
			Status:   getString(itemMap, "status"),
			Priority: getString(itemMap, "priority"),
		}
		if item.Status == "" {
			item.Status = "pending"
		}
		items = append(items, item)
	}

	GlobalTodoStore.mu.Lock()
	GlobalTodoStore.todos[sessionID] = items
	GlobalTodoStore.mu.Unlock()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Task list updated (%d items):\n", len(items)))
	for _, item := range items {
		statusChar := "◌"
		switch item.Status {
		case "completed":
			statusChar = "✓"
		case "in_progress":
			statusChar = "○"
		case "cancelled":
			statusChar = "✗"
		}
		priority := ""
		if item.Priority == "high" {
			priority = " 🔥"
		}
		result.WriteString(fmt.Sprintf("  %s %s%s\n", statusChar, item.Content, priority))
	}

	return result.String(), nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
	return ""
}

func GetTodos(sessionID string) []TodoItem {
	GlobalTodoStore.mu.RLock()
	defer GlobalTodoStore.mu.RUnlock()
	items := GlobalTodoStore.todos[sessionID]
	if items == nil {
		return []TodoItem{}
	}
	return items
}
