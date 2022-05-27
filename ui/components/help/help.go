package help

import (
	baseHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"github.com/coloradocolby/gh-eco/utils"
)

type Model struct {
	keys utils.KeyMap
	help baseHelp.Model
	ctx  *context.ProgramContext
}

func NewModel() Model {
	return Model{
		keys: utils.Keys,
		help: baseHelp.NewModel(),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, utils.Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	return styles.Subtle.Render(m.help.View(utils.Keys))
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}
