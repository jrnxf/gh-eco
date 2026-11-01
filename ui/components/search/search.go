package search

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
)

type Model struct {
	keys      keys.KeyMap
	textInput textinput.Model
	state     *state.State
}

func NewModel(s *state.State) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "search by username"

	return Model{
		keys:      keys.Keys,
		textInput: ti,
		state:     s,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state.Mode {
		case state.InsertMode:
			if login := strings.TrimSpace(m.textInput.Value()); key.Matches(msg, m.keys.Search) && login != "" {
				m.state.Mode = state.NormalMode
				cmds = append(cmds,
					m.textInput.Cursor.SetMode(cursor.CursorHide),
					func() tea.Msg { return commands.SearchSubmitted{Login: login} },
				)
			}

		case state.NormalMode:
			if key.Matches(msg, m.keys.FocusInput) && m.state.CurrentView == state.UserView {
				m.textInput.Reset()
				m.state.Mode = state.InsertMode
				return m, m.textInput.Cursor.SetMode(cursor.CursorBlink)
			}
		}
	}

	var cmd tea.Cmd
	if m.state.Mode == state.InsertMode {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return "\n" + m.textInput.View()
}
