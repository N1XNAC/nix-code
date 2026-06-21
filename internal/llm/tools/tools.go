package tools

import (
	"context"

	"github.com/n1xcode/n1x/internal/llm/provider"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Required() []string
	Execute(ctx context.Context, args map[string]any, sessionID string) (string, error)
}

func ToToolInfos(tools []Tool) []provider.ToolInfo {
	var result []provider.ToolInfo
	for _, t := range tools {
		result = append(result, provider.ToolInfo{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
			Required:    t.Required(),
		})
	}
	return result
}

func FindTool(tools []Tool, name string) Tool {
	for _, t := range tools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}
