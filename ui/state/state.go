// Package state holds the program state shared by the root model and its
// components. The root model allocates one State and hands every component
// the same pointer.
package state

import "github.com/jrnxf/gh-eco/models"

// RightSlack keeps every content line this many cells short of the screen
// edge. Some emoji draw wider than lipgloss measures them; a line that
// reached the last column would then wrap and shift the whole frame.
const RightSlack = 2

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type View int

const (
	UserView View = iota
	ReadmeView
	// PeopleView lists a user's followers or the people they follow.
	PeopleView
)

type TargetKind int

const (
	NoTarget TargetKind = iota
	UserTarget
	// FollowersTarget and FollowingTarget are a user's two counts, which
	// open the lists behind them.
	FollowersTarget
	FollowingTarget
	RepoTarget
)

// Target is one focusable element of the user view.
type Target struct {
	Kind TargetKind
	// Repo indexes User.Repos when Kind is RepoTarget.
	Repo int
}

// People is an open followers or following list.
type People struct {
	Kind models.PeopleKind
	// Login is whose list it is.
	Login string
	List  []models.Person
	Total int
	// Loaded is set once the first page has arrived, Failed once a page
	// failed to. Next continues the list while HasNext is set.
	Loaded  bool
	Failed  bool
	HasNext bool
	Next    string
	// Selected indexes List; Offset is the first visible entry.
	Selected int
	Offset   int
}

// Move moves the selection by delta entries, stopping at either end.
func (p *People) Move(delta int) {
	p.Selected = max(0, min(p.Selected+delta, len(p.List)-1))
}

// SelectedPerson returns the selected entry, or nil when the list is empty.
func (p *People) SelectedPerson() *models.Person {
	if p.Selected < 0 || p.Selected >= len(p.List) {
		return nil
	}
	return &p.List[p.Selected]
}

// Page is a profile as it was left: its view, focus and any open list.
type Page struct {
	User   models.User
	View   View
	Focus  int
	People People
}

type State struct {
	User        models.User
	CurrentView View
	Mode        Mode
	// Focus indexes Targets().
	Focus int
	// People is the list PeopleView shows.
	People People
	// History holds the pages left behind, most recent last.
	History []Page
	Layout  struct {
		ScreenHeight  int
		ScreenWidth   int
		ContentHeight int
		ContentWidth  int
	}
}

// Targets lists the focusable elements in focus order: nothing, the user,
// a user's follower and following counts, then each repo.
func (s *State) Targets() []Target {
	targets := []Target{{Kind: NoTarget}}
	if s.User.Login == "" {
		return targets
	}
	targets = append(targets, Target{Kind: UserTarget})
	if !s.User.IsOrganization {
		targets = append(targets, Target{Kind: FollowersTarget}, Target{Kind: FollowingTarget})
	}
	for i := range s.User.Repos {
		targets = append(targets, Target{Kind: RepoTarget, Repo: i})
	}
	return targets
}

func (s *State) Focused() Target {
	targets := s.Targets()
	if s.Focus < 0 || s.Focus >= len(targets) {
		return targets[0]
	}
	return targets[s.Focus]
}

// FocusedRepo returns the focused repo, or nil when focus is
// elsewhere. It points into User.Repos, so it always reflects the
// latest star state.
func (s *State) FocusedRepo() *models.Repo {
	if t := s.Focused(); t.Kind == RepoTarget {
		return &s.User.Repos[t.Repo]
	}
	return nil
}

// Repo returns the shown repo with the given id, or nil.
func (s *State) Repo(id string) *models.Repo {
	for i := range s.User.Repos {
		if s.User.Repos[i].Id == id {
			return &s.User.Repos[i]
		}
	}
	return nil
}

// Repos returns every copy of the repo with the given id: on screen and
// on the pages in history.
func (s *State) Repos(id string) []*models.Repo {
	var repos []*models.Repo
	collect := func(u *models.User) {
		for i := range u.Repos {
			if u.Repos[i].Id == id {
				repos = append(repos, &u.Repos[i])
			}
		}
	}
	collect(&s.User)
	for i := range s.History {
		collect(&s.History[i].User)
	}
	return repos
}

// FocusNext moves focus to the next target, wrapping from the last repo
// back to the profile. The empty target only exists before the first
// move: cycling never lands on it.
func (s *State) FocusNext() {
	n := len(s.Targets())
	if n == 1 {
		s.Focus = 0
		return
	}
	s.Focus = s.Focus%(n-1) + 1
}

// FocusPrev moves focus to the previous target, wrapping from the profile
// to the last repo.
func (s *State) FocusPrev() {
	n := len(s.Targets())
	switch {
	case n == 1:
		s.Focus = 0
	case s.Focus <= 1:
		s.Focus = n - 1
	default:
		s.Focus--
	}
}

// Push saves the page on screen so Pop can return to it. Nothing is saved
// while no profile is loaded, such as after a failed search. A page left
// from a readme returns to the profile under it.
func (s *State) Push() {
	if s.User.Login == "" {
		return
	}
	view := s.CurrentView
	if view == ReadmeView {
		view = UserView
	}
	s.History = append(s.History, Page{User: s.User, View: view, Focus: s.Focus, People: s.People})
}

// Pop returns to the most recently saved page. It reports false, changing
// nothing, when there is none.
func (s *State) Pop() bool {
	if len(s.History) == 0 {
		return false
	}
	p := s.History[len(s.History)-1]
	s.History = s.History[:len(s.History)-1]
	s.User, s.CurrentView, s.Focus, s.People = p.User, p.View, p.Focus, p.People
	return true
}
