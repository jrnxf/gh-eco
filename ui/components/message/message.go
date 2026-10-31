package message

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
)

type Model struct {
	keys    keys.KeyMap
	Content string
	timer   timer.Model
	state   *state.State
}

func NewModel(s *state.State) Model {
	return Model{
		keys:  keys.Keys,
		state: s,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {

	case commands.SetMessage:
		if m.state.CurrentView != state.MessageView {
			m.state.LastView = m.state.CurrentView
		}
		m.state.CurrentView = state.MessageView
		m.Content = msg.Content
		m.timer = timer.NewWithInterval(time.Second*time.Duration(msg.SecondsDisplayed), time.Second)
		cmd = m.timer.Init()

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.state.CurrentView = m.state.LastView
		m.state.LastView = state.VoidView
		m.Content = ""
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	text := lipgloss.NewStyle().Width(m.state.Layout.ScreenWidth).PaddingTop(m.state.Layout.ScreenHeight / 2).Align(lipgloss.Center).Render(
		// fmt.Sprintf("%s %s", m.Content, m.timer.View()),
		m.Content,
	)
	return text
}

func (m Model) TriggerMessage(content string, secondsDisplayed int) tea.Cmd {
	return func() tea.Msg {
		return commands.SetMessage{
			Content:          content,
			SecondsDisplayed: secondsDisplayed,
		}
	}
}
