package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadTool struct{}

func NewReadTool() *ReadTool {
	return &ReadTool{}
}

func (t *ReadTool) Name() string {
	return "read"
}

func (t *ReadTool) Description() string {
	return "Read the contents of a file. Can read specific line ranges for large files."
}

func (t *ReadTool) Parameters() map[string]any {
	return map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The path to the file to read",
		},
		"offset": map[string]any{
			"type":        "integer",
			"description": "The line number to start reading from (0-indexed)",
		},
		"limit": map[string]any{
			"type":        "integer",
			"description": "The maximum number of lines to read",
		},
	}
}

func (t *ReadTool) Required() []string {
	return []string{"file_path"}
}

func (t *ReadTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	if !filepath.IsAbs(filePath) {
		cwd, _ := os.Getwd()
		filePath = filepath.Join(cwd, filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	limit := len(lines)
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if offset >= len(lines) {
		return "", fmt.Errorf("offset %d exceeds file length %d", offset, len(lines))
	}

	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}

	selectedLines := lines[offset:end]
	totalLines := len(lines)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("File: %s (%d lines total, showing %d-%d)\n\n", filePath, totalLines, offset+1, end))
	for i, line := range selectedLines {
		result.WriteString(fmt.Sprintf("%d: %s\n", offset+i+1, line))
	}
	if end < totalLines {
		result.WriteString(fmt.Sprintf("\n... (%d more lines)\n", totalLines-end))
	}

	return result.String(), nil
}
