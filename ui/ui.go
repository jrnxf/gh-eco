package ui

import (
	"context"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cliBrowser "github.com/cli/browser"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/help"
	"github.com/jrnxf/gh-eco/ui/components/markdown"
	"github.com/jrnxf/gh-eco/ui/components/message"
	"github.com/jrnxf/gh-eco/ui/components/search"
	"github.com/jrnxf/gh-eco/ui/components/user"
	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
)

// Client is the GitHub API the UI depends on.
type Client interface {
	User(ctx context.Context, login string) (models.User, error)
	Readme(ctx context.Context, owner, name string) (string, error)
	OrgReadme(ctx context.Context, org string) (string, error)
	Star(ctx context.Context, id string) (models.Starrable, error)
	Unstar(ctx context.Context, id string) (models.Starrable, error)
}

type Model struct {
	client   Client
	state    *state.State
	keys     keys.KeyMap
	search   search.Model
	user     user.Model
	markdown markdown.Model
	help     help.Model
	message  message.Model
	// userReq identifies the latest user lookup.
	userReq int
}

func New(client Client) Model {
	s := &state.State{Mode: state.InsertMode, CurrentView: state.UserView}
	return Model{
		client:   client,
		state:    s,
		keys:     keys.Keys,
		search:   search.NewModel(s),
		user:     user.NewModel(s),
		markdown: markdown.NewModel(s),
		help:     help.NewModel(s),
		message:  message.NewModel(s),
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	s := m.state

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		if s.Mode == state.NormalMode {
			cmds = append(cmds, m.handleNormalKey(msg))
		}

	case commands.SearchSubmitted:
		m.userReq++
		cmds = append(cmds, fetchUser(m.client, msg.Login, m.userReq))

	case commands.GetUserResponse:
		if msg.Req != m.userReq {
			// a newer search superseded this one
			return m, nil
		}
		// a failed search clears the previous user, so nothing acts on it
		s.User = msg.User
		s.Focus = 0

	case commands.GetReadmeResponse:
		if msg.Err != nil {
			cmds = append(cmds, message.Show("couldn't load readme 😕", 2))
		} else {
			s.CurrentView = state.ReadmeView
		}

	case commands.StarResponse:
		if msg.Err != nil {
			text := "star failed 😕"
			if msg.Unstar {
				text = "unstar failed 😕"
			}
			cmds = append(cmds, message.Show(text, 2))
			break
		}
		if msg.Starrable.Id == github.GhEcoRepoID && !msg.Unstar {
			cmds = append(cmds, message.Show("tysm 🥹", 2))
		}
		for i := range s.User.PinnedRepos {
			r := &s.User.PinnedRepos[i]
			if r.Id == msg.Starrable.Id {
				r.ViewerHasStarred = msg.Starrable.ViewerHasStarred
				r.StarsCount = msg.Starrable.StargazerCount
			}
		}

	case tea.WindowSizeMsg:
		s.Layout.ScreenWidth = msg.Width
		s.Layout.ScreenHeight = msg.Height
	}

	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)
	m.search, cmd = m.search.Update(msg)
	cmds = append(cmds, cmd)
	m.message, cmd = m.message.Update(msg)
	cmds = append(cmds, cmd)
	// size the content area after every view change above, before the
	// components that fill it
	m.layout()
	m.markdown, cmd = m.markdown.Update(msg)
	cmds = append(cmds, cmd)
	m.user, cmd = m.user.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) handleNormalKey(msg tea.KeyMsg) tea.Cmd {
	s := m.state
	target := s.Focused()

	switch {
	case key.Matches(msg, m.keys.OpenGithub):
		switch target.Kind {
		case state.UserTarget:
			return openURL(s.User.Url)
		case state.RepoTarget:
			return openURL(s.FocusedRepo().Url)
		}

	case key.Matches(msg, m.keys.StarRepo):
		if r := s.FocusedRepo(); r != nil {
			return star(m.client, r.Id, r.ViewerHasStarred)
		}

	case key.Matches(msg, m.keys.StarGhEco):
		return star(m.client, github.GhEcoRepoID, false)

	case key.Matches(msg, m.keys.ToggleReadme):
		switch s.CurrentView {
		case state.UserView:
			switch target.Kind {
			case state.UserTarget:
				return fetchProfileReadme(m.client, s.User)
			case state.RepoTarget:
				r := s.FocusedRepo()
				return fetchReadme(m.client, r.Owner.Login, r.Name)
			}
		case state.ReadmeView:
			s.CurrentView = state.UserView
		}
	}

	if s.CurrentView == state.UserView {
		switch {
		case key.Matches(msg, m.keys.FocusNext):
			s.FocusNext()
		case key.Matches(msg, m.keys.FocusPrev):
			s.FocusPrev()
		case key.Matches(msg, m.keys.FocusInput):
			s.Focus = 0
		}
	}
	return nil
}

// layout sizes the content area between the search bar and the help bar.
func (m *Model) layout() {
	l := &m.state.Layout
	height := l.ScreenHeight - lipgloss.Height(m.help.View())
	if m.state.CurrentView == state.UserView {
		height -= lipgloss.Height(m.search.View())
	}
	l.ContentWidth = l.ScreenWidth
	l.ContentHeight = max(0, height)
}

func (m Model) View() string {
	// a notice takes the help bar's row for a moment
	bottom := m.help.View()
	if m.message.Content != "" {
		bottom = m.message.View()
	}

	if m.state.CurrentView == state.ReadmeView {
		return lipgloss.JoinVertical(lipgloss.Left, m.markdown.View(), bottom)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.search.View(), m.user.View(), bottom)
}

func fetchUser(c Client, login string, req int) tea.Cmd {
	return func() tea.Msg {
		u, err := c.User(context.Background(), login)
		return commands.GetUserResponse{Req: req, Login: login, User: u, Err: err}
	}
}

func fetchReadme(c Client, owner, name string) tea.Cmd {
	return func() tea.Msg {
		text, err := c.Readme(context.Background(), owner, name)
		return commands.GetReadmeResponse{Readme: text, Err: err}
	}
}

// fetchProfileReadme fetches the readme shown on a profile: a user's lives
// in the repo named after them, an organization's in its .github repo.
func fetchProfileReadme(c Client, u models.User) tea.Cmd {
	if !u.IsOrganization {
		return fetchReadme(c, u.Login, u.Login)
	}
	return func() tea.Msg {
		text, err := c.OrgReadme(context.Background(), u.Login)
		return commands.GetReadmeResponse{Readme: text, Err: err}
	}
}

// star stars id, or removes the star when starred is true.
func star(c Client, id string, starred bool) tea.Cmd {
	return func() tea.Msg {
		do := c.Star
		if starred {
			do = c.Unstar
		}
		s, err := do(context.Background(), id)
		return commands.StarResponse{Unstar: starred, Starrable: s, Err: err}
	}
}

func init() {
	// the default launcher would write into the terminal the UI owns
	cliBrowser.Stdout, cliBrowser.Stderr = io.Discard, io.Discard
}

// openURL opens url with the browser gh is configured to use.
func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		if err := browser.New("", io.Discard, io.Discard).Browse(url); err != nil {
			return commands.SetMessage{Content: "couldn't open the browser 😕", SecondsDisplayed: 2}
		}
		return nil
	}
}
