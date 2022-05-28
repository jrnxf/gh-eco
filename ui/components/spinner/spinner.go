package spinner

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	spinner spinner.Model
	Tick    tea.Cmd
}

func NewModel() Model {
	return Model{
		spinner: spinner.Model{
			Spinner: spinner.Dot,
			Style:   lipgloss.NewStyle().Foreground(lipgloss.Color("#5E81AC")),
		},
		Tick: spinner.Tick,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.spinner.View()
}
