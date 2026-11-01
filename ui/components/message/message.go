// Package message shows short notices in place of the help bar.
package message

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
)

var style = lipgloss.NewStyle().Bold(true).Foreground(styles.Accent)

type Model struct {
	Content string
	timer   timer.Model
	state   *state.State
}

func NewModel(s *state.State) Model {
	return Model{state: s}
}

// Show displays content for the given number of seconds.
func Show(content string, seconds int) tea.Cmd {
	return func() tea.Msg {
		return commands.SetMessage{Content: content, SecondsDisplayed: seconds}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.SetMessage:
		m.Content = msg.Content
		m.timer = timer.NewWithInterval(time.Second*time.Duration(msg.SecondsDisplayed), time.Second)
		return m, m.timer.Init()

	case tea.KeyMsg:
		// a keypress means the notice has been seen
		m.Content = ""

	case timer.TickMsg, timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		if msg.ID == m.timer.ID() {
			m.Content = ""
		}
	}
	return m, nil
}

// View renders the notice on one line, the height of the help bar it
// replaces.
func (m Model) View() string {
	return style.Render(ansi.Truncate(m.Content, m.state.Layout.ScreenWidth, "…"))
}
