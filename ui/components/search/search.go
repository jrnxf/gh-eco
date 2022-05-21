package search

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coloradocolby/ghx/api"
	"github.com/coloradocolby/ghx/ui/components/spinner"
	"github.com/coloradocolby/ghx/ui/context"
	"github.com/coloradocolby/ghx/utils"
)

type Model struct {
	keys      utils.KeyMap
	textInput textinput.Model
	spinner   spinner.Model
	fetching  bool
	err       error
	ctx       *context.ProgramContext
}

func NewModel() Model {
	ti := textinput.NewModel()
	ti.Focus()

	return Model{
		keys:      utils.Keys,
		textInput: ti,
		err:       nil,
		fetching:  false,
		spinner:   spinner.NewModel(),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd           tea.Cmd
		spinnerCmd    tea.Cmd
		textInputCmd  tea.Cmd
		searchUserCmd tea.Cmd

		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			cmd = tea.Quit
		}

		switch m.ctx.Mode {
		case context.InsertMode:
			cmds = append(cmds, textinput.Blink)

			if key.Matches(msg, m.keys.Search) {
				m.ctx.Mode = context.NormalMode
				m.textInput.SetCursorMode(textinput.CursorHide)
				searchUserCmd = api.SearchUser(m.textInput.Value())
				m.fetching = true
				cmds = append(cmds, m.spinner.Tick)
			}

		case context.NormalMode:
			if key.Matches(msg, m.keys.FocusInput) {
				m.textInput.Reset()
				m.ctx.Mode = context.InsertMode
				m.textInput.SetCursorMode(textinput.CursorBlink)
				return m, textinput.Blink
			}

		}
	case api.SearchUserResponse:
		m.fetching = false
	}

	m.textInput, textInputCmd = m.textInput.Update(msg)
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd, spinnerCmd, textInputCmd, searchUserCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.fetching {
		return fmt.Sprintf("%s%s\n", m.textInput.View(), m.spinner.View())
	} else {
		return m.textInput.View() + "\n"
	}
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}
