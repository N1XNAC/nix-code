package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/n1xcode/n1x/internal/config"
	"github.com/n1xcode/n1x/internal/llm/agent"
	"github.com/n1xcode/n1x/internal/permission"
	"github.com/n1xcode/n1x/internal/pubsub"
	"github.com/n1xcode/n1x/internal/tui"
	"github.com/n1xcode/n1x/internal/webserver"
)

type App struct {
	Config      *config.Config
	Agent       *agent.Agent
	Permissions *permission.PermissionService
	Bus         *pubsub.Bus
}

func New(cfg *config.Config) (*App, error) {
	bus := pubsub.NewBus()
	perms := permission.NewPermissionService()

	perms.LoadFromConfig(cfg.Permissions)
	perms.SetAgentPermissions(cfg.DefaultMode)

	if cfg.DefaultMode == "" {
		cfg.DefaultMode = "code"
	}

	a := &App{
		Config:      cfg,
		Bus:         bus,
		Permissions: perms,
	}

	ag, err := agent.New(cfg, bus, perms)
	if err != nil {
		return a, err
	}
	a.Agent = ag

	return a, nil
}

func (a *App) Shutdown() {
}

func (a *App) RunTUI(ctx context.Context) error {
	model := tui.InitialModel(ctx, a.Agent)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func (a *App) RunNonInteractive(ctx context.Context, prompt string) error {
	result, err := a.Agent.Run(ctx, prompt, "non-interactive")
	if err != nil {
		return fmt.Errorf("agent error: %w", err)
	}
	fmt.Println(result)
	return nil
}

func (a *App) RunConfigServer(ctx context.Context) error {
	srv := webserver.New(a.Config)
	return srv.Start(ctx)
}
