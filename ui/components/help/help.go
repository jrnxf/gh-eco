package help

import (
	baseHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
)

type Model struct {
	keys  keys.KeyMap
	help  baseHelp.Model
	state *state.State
}

func NewModel(s *state.State) Model {
	return Model{
		keys:  keys.Keys,
		help:  baseHelp.NewModel(),
		state: s,
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
	s := m.state
	bindings := []key.Binding{}
	switch s.Mode {
	case state.InsertMode:
		bindings = append(bindings, k.Search)
	case state.NormalMode:
		if s.User.Login == "" {
			// loading, or the search failed
			bindings = append(bindings, k.FocusInput)
		} else {
			if s.CurrentView == state.UserView {
				bindings = append(bindings, k.FocusInput, k.FocusNext, k.FocusPrev, k.ToggleReadme, k.OpenGithub)
				if s.Focused().Kind == state.RepoTarget {
					bindings = append(bindings, k.StarRepo)
				}
			} else if s.CurrentView == state.ReadmeView {
				bindings = append(bindings, k.FocusNext, k.FocusPrev, k.PreviewPageDown, k.PreviewPageUp, k.ToggleReadme, k.StarRepo, k.OpenGithub)
			}
			bindings = append(bindings, k.StarGhEco)
		}
	}
	bindings = append(bindings, k.Quit)

	return bindings
}
