package search

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/spinner"
	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/utils"
)

type Model struct {
	keys      keys.KeyMap
	textInput textinput.Model
	spinner   spinner.Model
	fetching  bool
	state     *state.State
}

func NewModel(s *state.State) Model {
	ti := textinput.NewModel()
	ti.Focus()
	ti.Placeholder = "search by username"

	return Model{
		keys:      keys.Keys,
		textInput: ti,
		spinner:   spinner.NewModel(),
		state:     s,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		spinnerCmd   tea.Cmd
		textInputCmd tea.Cmd

		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch m.state.Mode {
		case state.InsertMode:
			cmds = append(cmds, textinput.Blink)

			if login := strings.TrimSpace(m.textInput.Value()); key.Matches(msg, m.keys.Search) && login != "" {
				m.state.Mode = state.NormalMode
				m.textInput.SetCursorMode(textinput.CursorHide)
				m.fetching = true
				cmds = append(cmds, m.spinner.Tick, func() tea.Msg {
					return commands.SearchSubmitted{Login: login}
				})
			}

		case state.NormalMode:
			if key.Matches(msg, m.keys.FocusInput) {
				if m.state.CurrentView == state.UserView {
					m.textInput.Reset()
					m.state.Mode = state.InsertMode
					m.textInput.SetCursorMode(textinput.CursorBlink)
					return m, textinput.Blink
				}
			}
		}

	case commands.GetUserResponse:
		m.fetching = false
	}

	if m.state.Mode == state.InsertMode {
		m.textInput, textInputCmd = m.textInput.Update(msg)
	}

	if m.fetching {
		// by not sending updates to it i effectively stop the spinner and cut useless
		// re-renders in the top level update (in ui.go)
		m.spinner, spinnerCmd = m.spinner.Update(msg)
	}

	cmds = append(cmds, spinnerCmd, textInputCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder
	w := b.WriteString

	w(utils.GetNewLines(1))

	if m.fetching {
		w(fmt.Sprintf("%s%s", m.textInput.View(), m.spinner.View()))
	} else {
		w(m.textInput.View())
	}

	return lipgloss.NewStyle().Render(b.String())
}
