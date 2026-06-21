package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return Banner
	}

	var b strings.Builder

	header := m.renderHeader()
	b.WriteString(header + "\n")

	b.WriteString(m.renderViewport())

	b.WriteString("\n")

	b.WriteString(m.renderInput())

	b.WriteString("\n")

	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderHeader() string {
	modeText := ""
	modeStyle := lipgloss.NewStyle().Padding(0, 1)

	if m.mode == ModeThink {
		modeText = " Think "
		modeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#00BFFF")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)
	} else {
		modeText = " Code "
		modeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#4ecca3")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)
	}

	if m.width < 50 {
		title := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3")).Bold(true).Render("n1x")
		modelInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(string(m.agent.Provider().Model().Provider))
		return lipgloss.JoinHorizontal(lipgloss.Center,
			title,
			" ",
			modeStyle.Render(modeText),
			" ",
			modelInfo,
		)
	}

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3")).Bold(true).Render(" N1X Code ")
	providerName := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(string(m.agent.Provider().Model().Provider))
	modelName := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(m.agent.Provider().Model().ID)
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("─", max(0, m.width-50)))

	return lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		modeStyle.Render(modeText),
		" ",
		divider,
		" ",
		providerName,
		" ",
		modelName,
	)
}

func (m Model) renderViewport() string {
	if m.streaming && len(m.messages) > 0 {
		last := m.messages[len(m.messages)-1]
		if last.role == "assistant" {
			spinnerChar := m.spinner.View()
			last.content += "\n" + spinnerStyle.Render(spinnerChar)
		}
	}

	vp := m.viewport
	if m.width > 0 {
		vp.Width = m.width - 4
	}
	return vp.View()
}

func (m Model) renderInput() string {
	ta := m.textarea
	ta.SetWidth(m.width - 6)

	if !m.streaming {
		return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#333333")).Render(ta.View())
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4ecca3")).Render(ta.View() + "\n" + m.spinner.View() + " Thinking...")
}

func (m Model) renderStatusBar() string {
	modeLabel := ""
	if m.mode == ModeThink {
		modeLabel = lipgloss.NewStyle().
			Background(lipgloss.Color("#00BFFF")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1).
			Render(" Think ")
	} else {
		modeLabel = lipgloss.NewStyle().
			Background(lipgloss.Color("#4ecca3")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1).
			Render(" Code ")
	}

	statusText := ""
	if m.streaming {
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3")).Render(m.spinner.View() + " Working...")
	} else {
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Enter to send | Esc to navigate")
	}

	infoText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(fmt.Sprintf(" %d msgs", len(m.messages)))

	padding := m.width - lipgloss.Width(modeLabel) - lipgloss.Width(statusText) - lipgloss.Width(infoText) - 4
	if padding < 0 {
		padding = 0
	}

	barBg := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e"))
	return barBg.Render(
		modeLabel + statusText + strings.Repeat(" ", padding) + infoText,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ecca3"))
