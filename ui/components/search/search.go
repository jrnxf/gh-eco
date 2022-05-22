package search

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/ui/components/spinner"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/utils"
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

	// to save during dev time start with my username
	// ti.SetValue("coloradocolby")

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

	if m.ctx.Mode != context.NormalMode {
		m.textInput, textInputCmd = m.textInput.Update(msg)
	}

	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd, spinnerCmd, textInputCmd, searchUserCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder
	w := b.WriteString

	if m.fetching {
		w(fmt.Sprintf("%s%s", m.textInput.View(), m.spinner.View()))
	} else {
		w(m.textInput.View())
	}

	return lipgloss.NewStyle().Padding(1).Render(b.String())
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}