package ui

import (
	"context"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cliBrowser "github.com/cli/browser"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/help"
	"github.com/jrnxf/gh-eco/ui/components/loading"
	"github.com/jrnxf/gh-eco/ui/components/markdown"
	"github.com/jrnxf/gh-eco/ui/components/message"
	"github.com/jrnxf/gh-eco/ui/components/people"
	"github.com/jrnxf/gh-eco/ui/components/search"
	"github.com/jrnxf/gh-eco/ui/components/user"
	"github.com/jrnxf/gh-eco/ui/keys"
	"github.com/jrnxf/gh-eco/ui/state"
)

// Client is the GitHub API the UI depends on.
type Client interface {
	User(ctx context.Context, login string) (models.User, error)
	Viewer(ctx context.Context) (string, error)
	People(ctx context.Context, login string, kind models.PeopleKind, after string) (models.PeoplePage, error)
	Readme(ctx context.Context, owner, name string) (string, error)
	OrgReadme(ctx context.Context, org string) (string, error)
	CachedReadme(owner, name string) (string, bool)
	CachedOrgReadme(org string) (string, bool)
	Star(ctx context.Context, id string) (models.Starrable, error)
	Unstar(ctx context.Context, id string) (models.Starrable, error)
}

type Model struct {
	client   Client
	state    *state.State
	keys     keys.KeyMap
	search   search.Model
	user     user.Model
	people   people.Model
	markdown markdown.Model
	help     help.Model
	message  message.Model
	loading  loading.Model
	// login is the profile to open first, empty for the viewer's own.
	login string
	// userReq identifies the latest user lookup.
	userReq int
	// peopleReq identifies the latest followers or following page request;
	// peoplePending is set while it is unanswered.
	peopleReq     int
	peoplePending bool
	// readmeReq identifies the latest readme request; readmePending is set
	// while it is unanswered.
	readmeReq     int
	readmePending bool
	// starPending holds the repo ids with a star toggle in flight.
	starPending map[string]bool
}

// Labels for the loading spinner.
const (
	searching     = "searching…"
	loadingReadme = "loading readme…"
	loadingPeople = "loading…"
)

// New opens the profile of login, or the viewer's own when it is empty.
func New(client Client, login string) Model {
	s := &state.State{Mode: state.NormalMode, CurrentView: state.UserView}
	return Model{
		client:      client,
		state:       s,
		keys:        keys.Keys,
		login:       login,
		search:      search.NewModel(s),
		user:        user.NewModel(s),
		people:      people.NewModel(s),
		markdown:    markdown.NewModel(s),
		help:        help.NewModel(s),
		message:     message.NewModel(s),
		loading:     loading.New(),
		starPending: map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd {
	login := m.login
	return func() tea.Msg { return commands.SearchSubmitted{Login: login} }
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
			if m.readmePending {
				// any key means the user moved on: drop the pending
				// readme, success or failure
				m.cancelReadme()
			}
			cmds = append(cmds, m.handleNormalKey(msg))
		}

	case commands.SearchSubmitted:
		cmds = append(cmds, m.lookup(msg.Login))

	case commands.GetUserResponse:
		if msg.Req != m.userReq {
			// a newer search superseded this one
			return m, nil
		}
		m.loading.Stop(searching)
		// a readme or list opened on the old profile meanwhile no longer
		// applies
		m.cancelReadme()
		m.cancelPeople()
		// backspace returns to the page this one replaces
		s.Push()
		// a failed search clears the previous user, so nothing acts on it
		s.User = msg.User
		s.Focus = 0
		s.People = state.People{}
		// the searched profile replaces whatever was open on the old one
		s.CurrentView = state.UserView
		m.search.SetLogin(msg.Login)

	case commands.GetPeopleResponse:
		if msg.Req != m.peopleReq {
			// the list it was for was closed or replaced
			return m, nil
		}
		m.peoplePending = false
		m.loading.Stop(loadingPeople)
		p := &s.People
		if msg.Err != nil {
			p.Failed = true
			cmds = append(cmds, message.Show("couldn't load "+people.Noun(p.Kind)+" 😕", 2))
			break
		}
		p.List = append(p.List, msg.Page.People...)
		p.Total = msg.Page.TotalCount
		p.HasNext, p.Next = msg.Page.HasNextPage, msg.Page.EndCursor
		p.Loaded = true
		cmds = append(cmds, m.morePeople())

	case commands.GetReadmeResponse:
		if msg.Req != m.readmeReq {
			// a newer readme request superseded this one
			return m, nil
		}
		m.readmePending = false
		m.loading.Stop(loadingReadme)
		if s.Mode != state.NormalMode || s.CurrentView != state.UserView {
			// the user moved on before the readme arrived
			return m, nil
		}
		if msg.Err != nil {
			cmds = append(cmds, message.Show("couldn't load readme 😕", 2))
		} else {
			s.CurrentView = state.ReadmeView
		}

	case commands.StarResponse:
		delete(m.starPending, msg.Id)
		// the repo may be on screen or on a page in history
		for _, r := range s.Repos(msg.Id) {
			if msg.Err != nil {
				// undo the optimistic toggle
				setStarred(r, msg.Prev)
			} else {
				r.ViewerHasStarred = msg.Starrable.ViewerHasStarred
				r.StarsCount = msg.Starrable.StargazerCount
			}
		}
		if msg.Err != nil {
			text := "star failed: "
			if msg.Unstar {
				text = "unstar failed: "
			}
			cmds = append(cmds, message.Show(text+msg.Err.Error(), 3))
		}

	case tea.WindowSizeMsg:
		s.Layout.ScreenWidth = msg.Width
		s.Layout.ScreenHeight = msg.Height
	}

	var cmd tea.Cmd
	m.loading, cmd = m.loading.Update(msg)
	cmds = append(cmds, cmd)
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
	m.people, cmd = m.people.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// lookup fetches the profile of login, or the viewer's own when it is
// empty.
func (m *Model) lookup(login string) tea.Cmd {
	m.userReq++
	m.cancelReadme()
	return tea.Batch(fetchUser(m.client, login, m.userReq), m.loading.Start(searching))
}

// cancelReadme drops the pending readme request, if any.
func (m *Model) cancelReadme() {
	m.readmeReq++
	m.readmePending = false
	m.loading.Stop(loadingReadme)
}

// cancelPeople drops the pending list page request, if any.
func (m *Model) cancelPeople() {
	m.peopleReq++
	m.peoplePending = false
	m.loading.Stop(loadingPeople)
}

// back returns to the previous page, dropping any request still out for
// the one being left.
func (m *Model) back() tea.Cmd {
	s := m.state
	if !s.Pop() {
		return nil
	}
	m.userReq++
	m.loading.Stop(searching)
	m.cancelReadme()
	m.cancelPeople()
	m.user.Reset()
	m.search.SetLogin(s.User.Login)
	// a list left while its first page was out picks up where it was
	s.People.Failed = false
	return m.morePeople()
}

// openPeople opens the followers or following list of the profile on
// screen.
func (m *Model) openPeople(kind models.PeopleKind) tea.Cmd {
	s := m.state
	s.Push()
	total := s.User.FollowersCount
	if kind == models.Following {
		total = s.User.FollowingCount
	}
	s.People = state.People{Kind: kind, Login: s.User.Login, Total: total}
	s.CurrentView = state.PeopleView
	m.cancelPeople()
	return m.morePeople()
}

// morePeople fetches the next page of the open list once the selection
// is within a screenful of its end, or its first page if none arrived.
func (m *Model) morePeople() tea.Cmd {
	s := m.state
	p := s.People
	if s.CurrentView != state.PeopleView || m.peoplePending || p.Failed {
		return nil
	}
	if p.Loaded && (!p.HasNext || p.Selected < len(p.List)-s.Layout.ContentHeight) {
		return nil
	}
	m.peopleReq++
	m.peoplePending = true
	req, c := m.peopleReq, m.client
	fetch := func() tea.Msg {
		page, err := c.People(context.Background(), p.Login, p.Kind, p.Next)
		return commands.GetPeopleResponse{Req: req, Page: page, Err: err}
	}
	return tea.Batch(fetch, m.loading.Start(loadingPeople))
}

func (m *Model) handleNormalKey(msg tea.KeyMsg) tea.Cmd {
	s := m.state
	target := s.Focused()

	switch {
	case key.Matches(msg, m.keys.Back):
		return m.back()

	case key.Matches(msg, m.keys.OpenGithub):
		if s.CurrentView == state.PeopleView {
			if p := s.People.SelectedPerson(); p != nil {
				return openURL(p.Url)
			}
			return nil
		}
		switch target.Kind {
		case state.UserTarget:
			return openURL(s.User.Url)
		case state.FollowersTarget:
			return openURL(s.User.Url + "?tab=followers")
		case state.FollowingTarget:
			return openURL(s.User.Url + "?tab=following")
		case state.RepoTarget:
			return openURL(s.FocusedRepo().Url)
		}

	case key.Matches(msg, m.keys.StarRepo):
		if r := s.FocusedRepo(); r != nil && !m.starPending[r.Id] {
			// show the new state at once; a failure flips it back
			unstar := r.ViewerHasStarred
			setStarred(r, !unstar)
			m.starPending[r.Id] = true
			return star(m.client, r.Id, unstar, unstar)
		}

	case key.Matches(msg, m.keys.StarGhEco):
		// thank at once and show the star if gh-eco is pinned here; a
		// failure replaces the thanks and restores the card
		if m.starPending[github.GhEcoRepoID] {
			break
		}
		prev := false
		if r := s.Repo(github.GhEcoRepoID); r != nil {
			prev = r.ViewerHasStarred
			setStarred(r, true)
		}
		m.starPending[github.GhEcoRepoID] = true
		return tea.Batch(
			message.Show("tysm 🥹", 2),
			star(m.client, github.GhEcoRepoID, false, prev),
		)

	case key.Matches(msg, m.keys.ToggleReadme):
		switch s.CurrentView {
		case state.UserView:
			switch target.Kind {
			case state.UserTarget:
				return m.openReadme(profileReadme(s.User))
			case state.RepoTarget:
				r := s.FocusedRepo()
				return m.openReadme(readmeRef{owner: r.Owner.Login, name: r.Name})
			}
		case state.ReadmeView:
			s.CurrentView = state.UserView
		}
	}

	switch s.CurrentView {
	case state.UserView:
		switch {
		case key.Matches(msg, m.keys.FocusNext):
			s.FocusNext()
		case key.Matches(msg, m.keys.FocusPrev):
			s.FocusPrev()
		case key.Matches(msg, m.keys.FocusInput):
			s.Focus = 0
		case key.Matches(msg, m.keys.OpenList) && target.Kind == state.FollowersTarget:
			return m.openPeople(models.Followers)
		case key.Matches(msg, m.keys.OpenList) && target.Kind == state.FollowingTarget:
			return m.openPeople(models.Following)
		}

	case state.PeopleView:
		half := max(1, s.Layout.ContentHeight/2)
		switch {
		case key.Matches(msg, m.keys.FocusNext):
			s.People.Move(1)
		case key.Matches(msg, m.keys.FocusPrev):
			s.People.Move(-1)
		case key.Matches(msg, m.keys.PreviewPageDown):
			s.People.Move(half)
		case key.Matches(msg, m.keys.PreviewPageUp):
			s.People.Move(-half)
		case key.Matches(msg, m.keys.OpenProfile):
			if p := s.People.SelectedPerson(); p != nil {
				return m.lookup(p.Login)
			}
		}
		return m.morePeople()
	}
	return nil
}

// layout sizes the content area between the search bar and the help bar.
func (m *Model) layout() {
	l := &m.state.Layout
	height := l.ScreenHeight - lipgloss.Height(m.help.View())
	if m.state.CurrentView != state.ReadmeView {
		height -= lipgloss.Height(m.search.View())
	}
	l.ContentWidth = l.ScreenWidth
	l.ContentHeight = max(0, height)
}

func (m Model) View() string {
	// a notice, then a slow request's spinner, takes the help bar's row
	bottom := m.help.View()
	switch {
	case m.message.Content != "":
		bottom = m.message.View()
	case m.loading.Visible():
		bottom = m.loading.View(m.state.Layout.ScreenWidth)
	}

	// a plain join: lipgloss.JoinVertical pads every row to the widest
	// block, which would undo the user view's right slack
	switch m.state.CurrentView {
	case state.ReadmeView:
		return strings.Join([]string{m.markdown.View(), bottom}, "\n")
	case state.PeopleView:
		return strings.Join([]string{m.search.View(), m.people.View(), bottom}, "\n")
	}
	return strings.Join([]string{m.search.View(), m.user.View(), bottom}, "\n")
}

// fetchUser looks up login, or the viewer's own profile when it is empty.
func fetchUser(c Client, login string, req int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if login == "" {
			viewer, err := c.Viewer(ctx)
			if err != nil {
				return commands.GetUserResponse{Req: req, Err: err}
			}
			login = viewer
		}
		u, err := c.User(ctx, login)
		return commands.GetUserResponse{Req: req, Login: login, User: u, Err: err}
	}
}

// readmeRef names a readme: a repo's, or an organization's profile readme.
type readmeRef struct {
	owner, name string
	org         bool
}

// profileReadme is the readme shown on a profile: a user's lives in the
// repo named after them, an organization's in its .github repo.
func profileReadme(u models.User) readmeRef {
	if u.IsOrganization {
		return readmeRef{owner: u.Login, org: true}
	}
	return readmeRef{owner: u.Login, name: u.Login}
}

// openReadme requests a readme. A cached one answers at once; only a
// request that goes to the network shows the spinner.
func (m *Model) openReadme(ref readmeRef) tea.Cmd {
	m.readmeReq++
	m.readmePending = true
	req := m.readmeReq

	var text string
	var ok bool
	if ref.org {
		text, ok = m.client.CachedOrgReadme(ref.owner)
	} else {
		text, ok = m.client.CachedReadme(ref.owner, ref.name)
	}
	if ok {
		return func() tea.Msg { return commands.GetReadmeResponse{Req: req, Readme: text} }
	}

	c := m.client
	fetch := func() tea.Msg {
		var text string
		var err error
		if ref.org {
			text, err = c.OrgReadme(context.Background(), ref.owner)
		} else {
			text, err = c.Readme(context.Background(), ref.owner, ref.name)
		}
		return commands.GetReadmeResponse{Req: req, Readme: text, Err: err}
	}
	return tea.Batch(fetch, m.loading.Start(loadingReadme))
}

// star stars id, or removes the star when starred is true. prev is the
// state a failure restores.
func star(c Client, id string, starred, prev bool) tea.Cmd {
	return func() tea.Msg {
		do := c.Star
		if starred {
			do = c.Unstar
		}
		s, err := do(context.Background(), id)
		return commands.StarResponse{Id: id, Unstar: starred, Prev: prev, Starrable: s, Err: err}
	}
}

// setStarred sets the viewer's star on r and adjusts its count to match.
func setStarred(r *models.Repo, starred bool) {
	if r.ViewerHasStarred == starred {
		return
	}
	r.ViewerHasStarred = starred
	if starred {
		r.StarsCount++
	} else {
		r.StarsCount = max(0, r.StarsCount-1)
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
