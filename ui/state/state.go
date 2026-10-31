// Package state holds the program state shared by the root model and its
// components. The root model allocates one State and hands every component
// the same pointer.
package state

import "github.com/jrnxf/gh-eco/models"

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type View int

const (
	UserView View = iota
	ReadmeView
)

type TargetKind int

const (
	NoTarget TargetKind = iota
	UserTarget
	RepoTarget
)

// Target is one focusable element of the user view.
type Target struct {
	Kind TargetKind
	// Repo indexes User.PinnedRepos when Kind is RepoTarget.
	Repo int
}

type State struct {
	User        models.User
	CurrentView View
	Mode        Mode
	// Focus indexes Targets().
	Focus  int
	Layout struct {
		ScreenHeight  int
		ScreenWidth   int
		ContentHeight int
		ContentWidth  int
	}
}

// Targets lists the focusable elements in focus order: nothing, the user,
// then each pinned repo.
func (s *State) Targets() []Target {
	targets := []Target{{Kind: NoTarget}}
	if s.User.Login == "" {
		return targets
	}
	targets = append(targets, Target{Kind: UserTarget})
	for i := range s.User.PinnedRepos {
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

// FocusedRepo returns the focused pinned repo, or nil when focus is
// elsewhere. It points into User.PinnedRepos, so it always reflects the
// latest star state.
func (s *State) FocusedRepo() *models.Repo {
	if t := s.Focused(); t.Kind == RepoTarget {
		return &s.User.PinnedRepos[t.Repo]
	}
	return nil
}

func (s *State) FocusNext() {
	n := len(s.Targets())
	s.Focus = (s.Focus + 1) % n
}

func (s *State) FocusPrev() {
	n := len(s.Targets())
	s.Focus = (s.Focus - 1 + n) % n
}
