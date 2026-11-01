package state

import (
	"testing"

	"github.com/jrnxf/gh-eco/models"
)

func Test_FocusCycleSkipsTheEmptyTarget(t *testing.T) {
	s := &State{User: models.User{Login: "octocat", Repos: make([]models.Repo, 3)}}
	// targets: none, the profile, its two counts, then three repos

	var forward []TargetKind
	for range 7 {
		s.FocusNext()
		forward = append(forward, s.Focused().Kind)
	}
	want := []TargetKind{UserTarget, FollowersTarget, FollowingTarget, RepoTarget, RepoTarget, RepoTarget, UserTarget}
	for i := range want {
		if forward[i] != want[i] {
			t.Fatalf("j presses landed on %v, want %v", forward, want)
		}
	}

	var backward []TargetKind
	for range 7 {
		s.FocusPrev()
		backward = append(backward, s.Focused().Kind)
	}
	want = []TargetKind{RepoTarget, RepoTarget, RepoTarget, FollowingTarget, FollowersTarget, UserTarget, RepoTarget}
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

func Test_OrganizationsHaveNoCounts(t *testing.T) {
	s := &State{User: models.User{Login: "acme", IsOrganization: true, Repos: make([]models.Repo, 1)}}
	for _, target := range s.Targets() {
		if target.Kind == FollowersTarget || target.Kind == FollowingTarget {
			t.Fatalf("organization targets include %v", target.Kind)
		}
	}
}

func Test_HistoryPushPop(t *testing.T) {
	s := &State{}
	if s.Pop() {
		t.Fatal("popped with nothing in history")
	}

	// a failed search shows no profile, so there is nothing to return to
	s.Push()
	if len(s.History) != 0 {
		t.Fatalf("pushed a page with no profile: %+v", s.History)
	}

	s.User = models.User{Login: "alice"}
	s.Focus, s.CurrentView = 3, ReadmeView
	s.Push()

	s.User = models.User{Login: "alice"}
	s.CurrentView = PeopleView
	s.People = People{Login: "alice", List: []models.Person{{Login: "bob"}, {Login: "carol"}}, Selected: 1}
	s.Push()

	s.User, s.Focus, s.CurrentView, s.People = models.User{Login: "carol"}, 0, UserView, People{}

	if !s.Pop() || s.User.Login != "alice" || s.CurrentView != PeopleView || s.People.SelectedPerson().Login != "carol" {
		t.Fatalf("first pop = %q in view %v with %+v, want alice's list on carol", s.User.Login, s.CurrentView, s.People)
	}
	// a page left from a readme comes back as the profile, focus kept
	if !s.Pop() || s.User.Login != "alice" || s.CurrentView != UserView || s.Focus != 3 {
		t.Fatalf("second pop = %q in view %v, focus %d; want alice's profile, focus 3", s.User.Login, s.CurrentView, s.Focus)
	}
	if s.Pop() || s.User.Login != "alice" {
		t.Errorf("pop past the first page changed it to %q", s.User.Login)
	}
}
