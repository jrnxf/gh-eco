package state

import (
	"testing"

	"github.com/jrnxf/gh-eco/models"
)

func Test_FocusCycleSkipsTheEmptyTarget(t *testing.T) {
	s := &State{User: models.User{Login: "octocat", PinnedRepos: make([]models.Repo, 3)}}
	// targets: none, the profile, then three repos

	var forward []TargetKind
	for range 5 {
		s.FocusNext()
		forward = append(forward, s.Focused().Kind)
	}
	want := []TargetKind{UserTarget, RepoTarget, RepoTarget, RepoTarget, UserTarget}
	for i := range want {
		if forward[i] != want[i] {
			t.Fatalf("j presses landed on %v, want %v", forward, want)
		}
	}

	var backward []TargetKind
	for range 5 {
		s.FocusPrev()
		backward = append(backward, s.Focused().Kind)
	}
	want = []TargetKind{RepoTarget, RepoTarget, RepoTarget, UserTarget, RepoTarget}
	for i := range want {
		if backward[i] != want[i] {
			t.Fatalf("k presses landed on %v, want %v", backward, want)
		}
	}

	// k from nothing focused goes to the last repo
	s.Focus = 0
	s.FocusPrev()
	if t2 := s.Focused(); t2.Kind != RepoTarget || t2.Repo != 2 {
		t.Errorf("k from no focus landed on %+v, want the last repo", t2)
	}
}

func Test_FocusWithoutAUser(t *testing.T) {
	s := &State{}
	s.FocusNext()
	s.FocusPrev()
	if s.Focus != 0 {
		t.Errorf("focus moved to %d with nothing loaded", s.Focus)
	}
}
