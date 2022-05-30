package help

import (
	baseHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/coloradocolby/gh-eco/ui/context"
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
	}

	return m, nil
}

func (m Model) View() string {
	return m.help.ShortHelpView(m.collectHelpBindings())
}

func (m Model) collectHelpBindings() []key.Binding {
	k := m.keys
	bindings := []key.Binding{}
	switch m.ctx.Mode {
	case context.InsertMode:
		bindings = append(bindings, k.Search)
	case context.NormalMode:
		if m.ctx.User.Login == "" {
			// user has not yet loaded
			bindings = append(bindings, k.Search)
		} else {
			if m.ctx.View == context.UserView {
				fw := m.ctx.CurrentFocus.FocusedWidget
				bindings = append(bindings, k.FocusInput, k.FocusNext, k.FocusPrev, k.ToggleReadme, k.OpenGithub)
				if fw.Type == context.RepoWidget {
					bindings = append(bindings, k.StarRepo)
				} else if fw.Type == context.UserWidget {
					bindings = append(bindings, k.FollowUser)
				}
			} else if m.ctx.View == context.ReadmeView {
				bindings = append(bindings, k.FocusNext, k.FocusPrev, k.PreviewPageDown, k.PreviewPageUp, k.ToggleReadme, k.StarRepo, k.OpenGithub)
			}
		}
	}
	bindings = append(bindings, k.Quit)

	return bindings

}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}
