package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type EditTool struct{}

func NewEditTool() *EditTool {
	return &EditTool{}
}

func (t *EditTool) Name() string {
	return "edit"
}

func (t *EditTool) Description() string {
	return "Perform precise edits to files by replacing exact text matches. Uses multi-strategy matching to find the correct location."
}

func (t *EditTool) Parameters() map[string]any {
	return map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The path to the file to edit",
		},
		"old_string": map[string]any{
			"type":        "string",
			"description": "The text to replace",
		},
		"new_string": map[string]any{
			"type":        "string",
			"description": "The text to replace it with",
		},
	}
}

func (t *EditTool) Required() []string {
	return []string{"file_path", "old_string", "new_string"}
}

type replacer func(content, old, new string) (string, bool)

func (t *EditTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	filePath, _ := args["file_path"].(string)
	oldString, _ := args["old_string"].(string)
	newString, _ := args["new_string"].(string)

	if filePath == "" || oldString == "" {
		return "", fmt.Errorf("file_path and old_string are required")
	}

	if !filepath.IsAbs(filePath) {
		cwd, _ := os.Getwd()
		filePath = filepath.Join(cwd, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	contentStr := string(content)

	replacers := []replacer{
		exactMatch,
		trimTrailingWhitespace,
		trimLeadingTrailing,
		collapseWhitespace,
	}

	var result string
	var success bool
	for _, r := range replacers {
		result, success = r(contentStr, oldString, newString)
		if success {
			break
		}
	}

	if !success {
		matches := countOccurrences(contentStr, oldString)
		if matches > 0 {
			return "", fmt.Errorf("found %d match(es) but all replacement strategies failed. Please provide more context around the text to replace", matches)
		}
		return "", fmt.Errorf("text to replace not found in file. Verify the content and try again with more surrounding context")
	}

	if err := os.WriteFile(filePath, []byte(result), 0644); err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return fmt.Sprintf("File edited: %s", filePath), nil
}

func exactMatch(content, old, new string) (string, bool) {
	if strings.Contains(content, old) {
		if strings.Count(content, old) > 1 {
			return "", false
		}
		return strings.Replace(content, old, new, 1), true
	}
	return "", false
}

func trimTrailingWhitespace(content, old, new string) (string, bool) {
	oldTrimmed := strings.TrimRight(old, " \t\r\n")
	newTrimmed := strings.TrimRight(new, " \t\r\n")
	if strings.Count(content, oldTrimmed) == 1 {
		return strings.Replace(content, oldTrimmed, newTrimmed, 1), true
	}
	return "", false
}

func trimLeadingTrailing(content, old, new string) (string, bool) {
	oldTrimmed := strings.TrimSpace(old)
	newTrimmed := strings.TrimSpace(new)
	if strings.Count(content, oldTrimmed) == 1 {
		return strings.Replace(content, oldTrimmed, newTrimmed, 1), true
	}
	return "", false
}

func collapseWhitespace(content, old, new string) (string, bool) {
	collapse := func(s string) string {
		parts := strings.Fields(s)
		return strings.Join(parts, " ")
	}
	oldCollapsed := collapse(old)
	newCollapsed := collapse(new)
	if strings.Count(content, oldCollapsed) == 1 {
		return strings.Replace(content, oldCollapsed, newCollapsed, 1), true
	}
	return "", false
}

func countOccurrences(content, substr string) int {
	if substr == "" {
		return 0
	}
	count := 0
	for i := 0; i <= len(content)-len(substr); {
		if content[i:i+len(substr)] == substr {
			count++
			i += len(substr)
		} else {
			i++
		}
	}
	return count
}
