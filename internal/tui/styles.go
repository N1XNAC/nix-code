package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7B56DB"}
	Special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	Green     = lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}
	Red       = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
	Yellow    = lipgloss.AdaptiveColor{Light: "#FFFF00", Dark: "#FFFF00"}
	Blue      = lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#00BFFF"}
	Cyan      = lipgloss.AdaptiveColor{Light: "#00FFFF", Dark: "#00FFFF"}
	White     = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
)

var (
	BannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ecca3")).
			Bold(true).
			Align(lipgloss.Center).
			MarginTop(2).
			MarginBottom(1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ecca3")).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2a2a4a")).
			Padding(0, 1)

	ThinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00BFFF")).
			Bold(true).
			Padding(0, 1)

	CodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#4ecca3")).
			Bold(true).
			Padding(0, 1)

	MessageStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	UserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	ToolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 1)

	DocStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)
