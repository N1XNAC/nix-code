package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/n1xcode/n1x/internal/llm/provider"
)

type streamMsg struct {
	event provider.ProviderEvent
}

type streamDoneMsg struct {
	err error
}

type errMsg struct {
	err error
}

func submitPrompt(m *Model) tea.Cmd {
	prompt := strings.TrimSpace(m.textarea.Value())
	if prompt == "" {
		return nil
	}

	m.AddMessage("user", prompt)
	m.textarea.Reset()
	m.streaming = true
	m.streamBuf.Reset()

	eventCh := make(chan provider.ProviderEvent, 100)

	go func() {
		m.agent.StreamRun(m.ctx, prompt, m.session, eventCh)
		close(eventCh)
	}()

	return func() tea.Msg {
		for event := range eventCh {
			switch event.Type {
			case provider.EventContentDelta:
				return streamMsg{event: event}
			case provider.EventToolUseStart:
				return streamMsg{event: event}
			case provider.EventToolUseDelta:
				return streamMsg{event: event}
			case provider.EventToolUseStop:
				args := map[string]any{}
				json.Unmarshal([]byte(event.Input), &args)
				desc := fmt.Sprintf("🔧 Using tool: %s", event.Name)
				if v, ok := args["file_path"]; ok {
					desc += fmt.Sprintf(" (%s)", v)
				}
				if v, ok := args["command"]; ok {
					desc += fmt.Sprintf(" (%s)", v)
				}
				m.AddToolMessage(desc)
			case provider.EventComplete:
				return streamDoneMsg{err: nil}
			case provider.EventError:
				return streamDoneMsg{err: event.Err}
			}
		}
		return streamDoneMsg{err: nil}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 2
		inputHeight := 5
		statusHeight := 1
		vpHeight := m.height - headerHeight - inputHeight - statusHeight - 2

		m.viewport.Width = m.width - 4
		m.viewport.Height = vpHeight
		m.textarea.SetWidth(m.width - 6)
		m.renderMessages()
		m.ready = true

	case tea.KeyMsg:
		if !m.textarea.Focused() {
			m.viewport, _ = m.viewport.Update(msg)
		}

		switch msg.String() {
		case "tab":
			m.ToggleMode()
			return m, nil

		case "ctrl+c":
			m.cancel()
			return m, tea.Quit

		case "enter":
			if !m.streaming && strings.TrimSpace(m.textarea.Value()) != "" {
				return m, tea.Batch(submitPrompt(&m), m.spinner.Tick)
			}

		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			} else {
				m.textarea.Focus()
			}

		case "up", "down":
			if !m.textarea.Focused() {
				m.viewport, _ = m.viewport.Update(msg)
			} else {
				m.textarea, _ = m.textarea.Update(msg)
			}
		}

	case streamMsg:
		e := msg.event
		switch e.Type {
		case provider.EventContentDelta:
			m.streamBuf.WriteString(e.Content)
			m.messages = m.messages[:len(m.messages)]
			lastMsg := messageItem{role: "assistant", content: m.streamBuf.String()}
			m.messages = append(m.messages, lastMsg)
			m.renderMessages()

		case provider.EventToolUseStart:
			m.AddToolMessage(fmt.Sprintf("  Running tool: %s...", e.Name))
		}
		return m, m.spinner.Tick

	case streamDoneMsg:
		m.streaming = false
		if m.streamBuf.Len() > 0 {
			m.AddMessage("assistant", m.streamBuf.String())
		}
		m.streamBuf.Reset()
		if msg.err != nil {
			m.err = msg.err
			m.AddMessage("tool", fmt.Sprintf("Error: %s", msg.err))
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.streaming = false
		return m, nil

	case spinner.TickMsg:
		if m.streaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	if !m.textarea.Focused() {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

