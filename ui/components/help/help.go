package help

import (
	baseHelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
)

type Model struct {
	keys  keys.KeyMap
	help  baseHelp.Model
	state *state.State
}

func NewModel(s *state.State) Model {
	return Model{
		keys:  keys.Keys,
		help:  newHelp(),
		state: s,
	}
}

func newHelp() baseHelp.Model {
	h := baseHelp.New()
	h.Styles.ShortKey = styles.HelpKey
	h.Styles.ShortDesc = styles.HelpDesc
	h.Styles.ShortSeparator = styles.HelpSeparator
	h.Styles.Ellipsis = styles.HelpSeparator
	return h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	}

	return m, nil
}

func (m Model) View() string {
	// ShortHelpView can overflow its width when the ellipsis does not fit
	return ansi.Truncate(m.help.ShortHelpView(m.collectHelpBindings()), m.help.Width, "…")
}

func (m Model) collectHelpBindings() []key.Binding {
	k := m.keys
	s := m.state
	bindings := []key.Binding{}
	switch s.Mode {
	case state.InsertMode:
		bindings = append(bindings, k.Search)
		if s.LookedUp {
			// esc leaves the search; ctrl+c still quits
			return append(bindings, k.Cancel)
		}
	case state.NormalMode:
		if s.User.Login == "" {
			// loading, or the search failed
			bindings = append(bindings, k.FocusInput)
			if len(s.History) > 0 {
				bindings = append(bindings, k.Back)
			}
		} else {
			switch s.CurrentView {
			case state.UserView:
				bindings = append(bindings, k.FocusInput, k.FocusNext, k.FocusPrev)
				if len(s.History) > 0 {
					bindings = append(bindings, k.Back)
				}
				switch s.Focused().Kind {
				case state.FollowersTarget, state.FollowingTarget:
					bindings = append(bindings, k.OpenList)
				default:
					bindings = append(bindings, k.ToggleReadme)
				}
				bindings = append(bindings, k.OpenGithub)
				if s.Focused().Kind == state.RepoTarget {
					bindings = append(bindings, k.StarRepo)
				}
			case state.PeopleView:
				bindings = append(bindings, k.FocusNext, k.FocusPrev, k.OpenProfile, k.Back, k.OpenGithub, k.FocusInput)
			case state.ReadmeView:
				bindings = append(bindings, k.FocusNext, k.FocusPrev, k.PreviewPageDown, k.PreviewPageUp, k.ToggleReadme, k.Back)
				if s.Focused().Kind == state.RepoTarget {
					// a profile readme has no repo to star
					bindings = append(bindings, k.StarRepo)
				}
				bindings = append(bindings, k.OpenGithub)
			}
			bindings = append(bindings, k.StarGhEco)
		}
	}
	bindings = append(bindings, k.Quit)

	return bindings
}
