package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type BashTool struct{}

func NewBashTool() *BashTool {
	return &BashTool{}
}

func (t *BashTool) Name() string {
	return "bash"
}

func (t *BashTool) Description() string {
	return "Execute shell commands in the terminal. Returns stdout, stderr, and exit code."
}

func (t *BashTool) Parameters() map[string]any {
	return map[string]any{
		"command": map[string]any{
			"type":        "string",
			"description": "The shell command to execute",
		},
		"workdir": map[string]any{
			"type":        "string",
			"description": "Working directory to run the command in",
		},
		"timeout": map[string]any{
			"type":        "integer",
			"description": "Timeout in milliseconds (default: 30000)",
		},
	}
}

func (t *BashTool) Required() []string {
	return []string{"command"}
}

func (t *BashTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	command, _ := args["command"].(string)
	workdir, _ := args["workdir"].(string)
	timeoutMs := 30000
	if t, ok := args["timeout"].(float64); ok {
		timeoutMs = int(t)
	}

	if command == "" {
		return "", fmt.Errorf("command is required")
	}

	shell := "sh"
	shellFlag := "-c"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "bash"
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, shellFlag, command)

	if workdir != "" {
		cmd.Dir = workdir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(fmt.Sprintf("STDOUT:\n%s\n", stdout.String()))
	}
	if stderr.Len() > 0 {
		result.WriteString(fmt.Sprintf("STDERR:\n%s\n", stderr.String()))
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			result.WriteString(fmt.Sprintf("\nCommand timed out after %dms", timeoutMs))
			return result.String(), nil
		} else {
			return "", fmt.Errorf("error executing command: %w", err)
		}
	}

	result.WriteString(fmt.Sprintf("\nExit code: %d", exitCode))
	if result.Len() == 0 {
		result.WriteString("(no output)")
	}

	return trimOutput(result.String(), 10000), nil
}

func trimOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + fmt.Sprintf("\n... (output truncated, %d bytes total)", len(s))
}
