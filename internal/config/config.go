package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	WorkingDir      string                `json:"wd,omitempty"`
	Providers       map[string]Provider   `json:"providers,omitempty"`
	DefaultProvider string                `json:"defaultProvider,omitempty"`
	DefaultMode     string                `json:"defaultMode,omitempty"`
	Theme           string                `json:"theme,omitempty"`
	AutoCompact     bool                  `json:"autoCompact,omitempty"`
	LSP         map[string]LSPConfig  `json:"lsp,omitempty"`
	MCPServers  map[string]MCPServer  `json:"mcpServers,omitempty"`
	Permissions map[string]string     `json:"permissions,omitempty"`
}

type Provider struct {
	APIKey  string `json:"apiKey,omitempty"`
	Model   string `json:"model,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type LSPConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type MCPServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

const (
	AppName        = "n1x"
	ConfigDirName  = "n1x"
	ConfigFileName = "config.json"
)

var cfg *Config

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", ConfigDirName), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func localConfigPath() string {
	return ".nix.json"
}

func Load(workingDir string, debug bool) (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		DefaultMode: "code",
		Theme:       "dark",
		AutoCompact: true,
		Providers:   make(map[string]Provider),
		LSP:         make(map[string]LSPConfig),
		MCPServers:  make(map[string]MCPServer),
		Permissions: map[string]string{
			"bash":  "ask",
			"write": "allow",
			"edit":  "allow",
		},
	}

	globalPath, err := configPath()
	if err != nil {
		return cfg, nil
	}

	if err := loadFromFile(globalPath, cfg); err != nil {
		return cfg, nil
	}

	if workingDir != "" {
		localPath := filepath.Join(workingDir, localConfigPath())
		if err := loadFromFile(localPath, cfg); err == nil {
			cfg.WorkingDir = workingDir
		}
	}

	return cfg, nil
}

func loadFromFile(path string, c *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *Config) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) GetProvider(name string) (Provider, bool) {
	p, ok := c.Providers[name]
	return p, ok
}

func (c *Config) SetProvider(name string, p Provider) {
	c.Providers[name] = p
}

func (c *Config) GetActiveProvider() (string, Provider, bool) {
	if c.DefaultProvider != "" {
		if p, ok := c.Providers[c.DefaultProvider]; ok && p.APIKey != "" {
			return c.DefaultProvider, p, true
		}
	}
	preferred := []string{"anthropic", "openai", "gemini", "openrouter", "groq", "deepseek", "mistral"}
	for _, name := range preferred {
		if p, ok := c.Providers[name]; ok && p.APIKey != "" {
			return name, p, true
		}
	}
	for name, p := range c.Providers {
		if p.APIKey != "" {
			return name, p, true
		}
	}
	return "", Provider{}, false
}

func WorkingDirectory() string {
	if cfg != nil && cfg.WorkingDir != "" {
		return cfg.WorkingDir
	}
	cwd, _ := os.Getwd()
	return cwd
}
