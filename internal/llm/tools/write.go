package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type WriteTool struct{}

func NewWriteTool() *WriteTool {
	return &WriteTool{}
}

func (t *WriteTool) Name() string {
	return "write"
}

func (t *WriteTool) Description() string {
	return "Create or overwrite a file with new content."
}

func (t *WriteTool) Parameters() map[string]any {
	return map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The path to the file to write",
		},
		"content": map[string]any{
			"type":        "string",
			"description": "The content to write to the file",
		},
	}
}

func (t *WriteTool) Required() []string {
	return []string{"file_path", "content"}
}

func (t *WriteTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)

	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	if !filepath.IsAbs(filePath) {
		cwd, _ := os.Getwd()
		filePath = filepath.Join(cwd, filePath)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	oldContent := ""
	if existing, err := os.ReadFile(filePath); err == nil {
		oldContent = string(existing)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	result := fmt.Sprintf("File written: %s (%d bytes)", filePath, len(content))
	if oldContent != "" {
		result += fmt.Sprintf(" (overwrote existing file, %d bytes)", len(oldContent))
	}
	return result, nil
}
