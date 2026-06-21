package tui

import (
	"encoding/json"
	"fmt"
	"strings"

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

func readNextEvent(ch chan provider.ProviderEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return streamDoneMsg{}
		}
		return streamMsg{event: event}
	}
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

	m.eventCh = make(chan provider.ProviderEvent, 100)

	go func() {
		m.agent.StreamRun(m.ctx, prompt, m.session, m.eventCh)
		close(m.eventCh)
	}()

	return readNextEvent(m.eventCh)
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
				return m, tea.Batch(submitPrompt(&m))
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
			return m, readNextEvent(m.eventCh)

		case provider.EventToolUseStart:
			m.AddToolMessage(fmt.Sprintf("  Running tool: %s...", e.Name))
			return m, readNextEvent(m.eventCh)

		case provider.EventToolUseDelta:
			return m, readNextEvent(m.eventCh)

		case provider.EventToolUseStop:
			args := map[string]any{}
			json.Unmarshal([]byte(e.Input), &args)
			desc := fmt.Sprintf("  Tool: %s", e.Name)
			if v, ok := args["file_path"]; ok {
				desc += fmt.Sprintf(" (%s)", v)
			}
			if v, ok := args["command"]; ok {
				desc += fmt.Sprintf(" (%s)", v)
			}
			m.AddToolMessage(desc)
			return m, readNextEvent(m.eventCh)

		case provider.EventComplete:
			return m, readNextEvent(m.eventCh)

		case provider.EventError:
			m.streaming = false
			m.AddMessage("tool", fmt.Sprintf("Error: %s", e.Err))
			return m, nil
		}

	case streamDoneMsg:
		m.streaming = false
		m.eventCh = nil
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
		m.eventCh = nil
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
