package textinput

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	TextInput textinput.Model
	err       error
}

func NewModel() Model {
	ti := textinput.NewModel()
	ti.Focus()
	ti.Prompt = "> "

	return Model{
		TextInput: ti,
		err:       nil,
	}
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.TextInput.View()
}

func (m *Model) HideCursor() tea.Cmd {
	return m.TextInput.SetCursorMode(textinput.CursorHide)
}
