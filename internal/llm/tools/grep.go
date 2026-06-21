package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GrepTool struct{}

func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

func (t *GrepTool) Name() string {
	return "grep"
}

func (t *GrepTool) Description() string {
	return "Search for a pattern in files using regex. Supports file pattern filtering."
}

func (t *GrepTool) Parameters() map[string]any {
	return map[string]any{
		"pattern": map[string]any{
			"type":        "string",
			"description": "The regex pattern to search for",
		},
		"include": map[string]any{
			"type":        "string",
			"description": "File pattern to include (e.g. *.go, *.{ts,tsx})",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "Directory to search in (default: current directory)",
		},
	}
}

func (t *GrepTool) Required() []string {
	return []string{"pattern"}
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	pattern, _ := args["pattern"].(string)
	include, _ := args["include"].(string)
	path, _ := args["path"].(string)

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	if path == "" {
		path = "."
	}

	if _, err := exec.LookPath("rg"); err == nil {
		return ripgrep(ctx, pattern, include, path)
	}

	return fallbackGrep(ctx, pattern, include, path)
}

func ripgrep(ctx context.Context, pattern, include, path string) (string, error) {
	args := []string{"--line-number", "--no-heading", "--color", "never", pattern, path}
	if include != "" {
		args = append([]string{"--glob", include}, args...)
	}

	cmd := exec.CommandContext(ctx, "rg", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if stdout.Len() == 0 {
				return "No matches found.", nil
			}
		} else {
			return "", fmt.Errorf("ripgrep error: %w", err)
		}
	}

	return trimOutput(stdout.String(), 10000), nil
}

func fallbackGrep(ctx context.Context, pattern, include, path string) (string, error) {
	args := []string{"-rn", "--color=never", pattern, path}
	if include != "" {
		args = append([]string{"--include", include}, args...)
	}

	cmd := exec.CommandContext(ctx, "grep", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if stdout.Len() == 0 {
				return "No matches found.", nil
			}
		} else {
			return "", fmt.Errorf("grep error: %w", err)
		}
	}

	result := stdout.String()
	lines := strings.Split(result, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
		result = strings.Join(lines, "\n")
		result += fmt.Sprintf("\n... (%d more lines)", len(strings.Split(stdout.String(), "\n"))-200)
	}

	return result, nil
}
