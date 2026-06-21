package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/n1xcode/n1x/internal/llm/agent"
	"github.com/n1xcode/n1x/internal/llm/provider"
)

type Mode int

const (
	ModeThink Mode = iota
	ModeCode
)

type messageItem struct {
	role    string
	content string
	toolUse bool
}

type Model struct {
	ctx        context.Context
	cancel     context.CancelFunc
	agent      *agent.Agent
	session    string
	ready      bool
	mode       Mode
	width      int
	height     int
	textarea   textarea.Model
	viewport   viewport.Model
	spinner    spinner.Model
	messages   []messageItem
	streaming  bool
	streamBuf  strings.Builder
	err        error
	eventCh    chan provider.ProviderEvent
}

func InitialModel(ctx context.Context, ag *agent.Agent) Model {
	ctx, cancel := context.WithCancel(ctx)
	ta := textarea.New()
	ta.Placeholder = "Ask me anything about your code... (Tab to toggle Think/Code mode)"
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(3)

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3"))
	s.Spinner = spinner.Dot

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return Model{
		ctx:        ctx,
		cancel:     cancel,
		agent:      ag,
		session:    fmt.Sprintf("session-%d", time.Now().UnixNano()),
		mode:     ModeCode,
		textarea: ta,
		viewport:   vp,
		spinner:    s,
		messages: []messageItem{
			{role: "system", content: Banner + "\n\nN1X Code is ready. Ask me about your code or tell me what to build."},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.spinner.Tick)
}

func (m *Model) AddMessage(role, content string) {
	m.messages = append(m.messages, messageItem{role: role, content: content})
	m.renderMessages()
}

func (m *Model) AddToolMessage(content string) {
	m.messages = append(m.messages, messageItem{role: "tool", content: content, toolUse: true})
	m.renderMessages()
}

func (m *Model) renderMessages() {
	var b strings.Builder
	for _, msg := range m.messages {
		switch msg.role {
		case "system":
			b.WriteString(msg.content + "\n\n")
		case "user":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Bold(true).Render("You:") + "\n")
			b.WriteString(msg.content + "\n\n")
		case "assistant":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3")).Bold(true).Render("N1X Code:") + "\n")
			b.WriteString(msg.content + "\n\n")
		case "tool":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true).Render("  " + msg.content) + "\n")
		}
	}
	m.viewport.SetContent(b.String())
	if m.viewport.ScrollPercent() > 0.9 {
		m.viewport.GotoBottom()
	}
}

func (m *Model) ToggleMode() {
	if m.mode == ModeThink {
		m.mode = ModeCode
		m.agent.SetMode(agent.ModeCode)
	} else {
		m.mode = ModeThink
		m.agent.SetMode(agent.ModeThink)
	}
}

func (m *Model) ModeLabel() string {
	if m.mode == ModeThink {
		return "Think"
	}
	return "Code"
}
