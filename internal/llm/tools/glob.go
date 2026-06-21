package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GlobTool struct{}

func NewGlobTool() *GlobTool {
	return &GlobTool{}
}

func (t *GlobTool) Name() string {
	return "glob"
}

func (t *GlobTool) Description() string {
	return "Search for files using glob patterns. Supports **, *, and ? wildcards."
}

func (t *GlobTool) Parameters() map[string]any {
	return map[string]any{
		"pattern": map[string]any{
			"type":        "string",
			"description": "The glob pattern to search for (e.g. **/*.go, src/**/*.ts)",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "Directory to search in (default: current directory)",
		},
	}
}

func (t *GlobTool) Required() []string {
	return []string{"pattern"}
}

func (t *GlobTool) Execute(ctx context.Context, args map[string]any, sessionID string) (string, error) {
	pattern, _ := args["pattern"].(string)
	searchPath, _ := args["path"].(string)

	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	if searchPath == "" {
		searchPath = "."
	}

	if !filepath.IsAbs(searchPath) {
		cwd, _ := os.Getwd()
		searchPath = filepath.Join(cwd, searchPath)
	}

	matches, err := filepath.Glob(filepath.Join(searchPath, pattern))
	if err != nil {
		return "", fmt.Errorf("invalid glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return "No files found matching pattern: " + pattern, nil
	}

	sort.Slice(matches, func(i, j int) bool {
		infoI, errI := os.Stat(matches[i])
		infoJ, errJ := os.Stat(matches[j])
		if errI == nil && errJ == nil {
			return infoI.ModTime().After(infoJ.ModTime())
		}
		return matches[i] < matches[j]
	})

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d file(s) matching '%s':\n\n", len(matches), pattern))
	for _, m := range matches {
		rel, _ := filepath.Rel(searchPath, m)
		info, err := os.Stat(m)
		if err == nil {
			result.WriteString(fmt.Sprintf("  %s (%d bytes, %s)\n", rel, info.Size(), info.ModTime().Format("Jan 02 15:04")))
		} else {
			result.WriteString(fmt.Sprintf("  %s\n", rel))
		}
	}

	return result.String(), nil
}
