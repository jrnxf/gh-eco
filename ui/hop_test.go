package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

var backspace = tea.KeyMsg{Type: tea.KeyBackspace}

// launch starts the program as main does with a login, and delivers the
// first lookup.
func launch(t *testing.T, client *fakeClient, login string, width, height int) Model {
	t.Helper()
	m, _ := send(New(client, login), tea.WindowSizeMsg{Width: width, Height: height})
	m, _ = send(m, find[commands.GetUserResponse](t, run(m.Init())))
	return m
}

// manyPeople makes n people named prefix-0 onwards.
func manyPeople(prefix string, n int) []models.Person {
	var list []models.Person
	for i := range n {
		login := fmt.Sprintf("%s-%d", prefix, i)
		list = append(list, models.Person{Login: login, Name: "Name " + login, Url: "https://github.com/" + login})
	}
	return list
}

// openList focuses the followers or following count and presses enter,
// delivering the first page.
func openList(t *testing.T, m Model, target state.TargetKind) Model {
	t.Helper()
	for m.state.Focused().Kind != target {
		m, _ = send(m, keyRunes("j"))
	}
	m, msgs := send(m, enter)
	if m.state.CurrentView != state.PeopleView {
		t.Fatalf("enter on %v opened view %v", target, m.state.CurrentView)
	}
	m, _ = send(m, find[commands.GetPeopleResponse](t, msgs))
	return m
}

func Test_LaunchOpensTheGivenLogin(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := launch(t, client, "octocat", 80, 24)

	if m.state.User.Login != "octocat" || m.state.Mode != state.NormalMode {
		t.Fatalf("launch with octocat shows %q in mode %v", m.state.User.Login, m.state.Mode)
	}
	if trail, search := topRows(m); trail != "octocat" || search != "" {
		t.Errorf("top rows are %q and %q, want the trail and no search line", trail, search)
	}

	// / still searches
	m, res := lookup(t, m, "octocat")
	if m, _ = send(m, res); m.state.User.Login != "octocat" {
		t.Errorf("search after launch failed")
	}
}

func Test_LaunchShowsTheLoginLoadingFromTheFirstFrame(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m, _ := send(New(client, "octocat"), tea.WindowSizeMsg{Width: 80, Height: 24})

	if trail, search := topRows(m); trail != "octocat" || search != "" {
		t.Errorf("first frame top rows are %q and %q, want the trail and no search line", trail, search)
	}
	if !spinnerShows(m, "loading octocat…") {
		t.Errorf("first frame does not show octocat loading:\n%s", m.View())
	}
}

func Test_LaunchWithUnknownLoginShowsNotFound(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{}}
	m := launch(t, client, "ghost", 80, 24)

	if !strings.Contains(m.View(), `no user or organization named "ghost"`) {
		t.Errorf("unknown login does not show the failed search state:\n%s", m.View())
	}
	m, _ = send(m, backspace)
	if m.state.User.Login != "" {
		t.Errorf("backspace on the first page loaded %q", m.state.User.Login)
	}
}

func Test_LaunchWithoutLoginStartsInTheSearchBox(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: 80, Height: 24})
	if _, ok := firstOf[commands.GetUserResponse](run(m.Init()).all(t)); ok {
		t.Errorf("launch without a login looked someone up")
	}
	if m.state.Mode != state.InsertMode || len(client.calls) > 0 {
		t.Fatalf("launch without a login is in mode %v after calls %v", m.state.Mode, client.calls)
	}
	if _, search := topRows(m); !strings.Contains(search, "search by username") {
		t.Errorf("search box is not showing: %q", search)
	}

	m, res := lookup(t, m, "octocat")
	if m, _ = send(m, res); m.state.User.Login != "octocat" {
		t.Errorf("search from launch failed")
	}
}

func Test_HopThroughFollowersAndBack(t *testing.T) {
	alice := models.User{Login: "alice", Url: "https://github.com/alice", FollowersCount: 3, Repos: []models.Repo{{Id: "R_A", Name: "alpha"}}}
	client := &fakeClient{
		users: map[string]models.User{"alice": alice, "bob": {Login: "bob", FollowingCount: 1}},
		people: map[string][]models.Person{
			"followers alice": {{Login: "carol"}, {Login: "bob", Name: "Bob B"}, {Login: "dave"}},
			"following bob":   {{Login: "alice"}},
		},
	}
	m := launch(t, client, "alice", 80, 24)

	m = openList(t, m, state.FollowersTarget)
	view := m.View()
	for _, want := range []string{"3 followers", "carol", "Bob B", "dave"} {
		if !strings.Contains(view, want) {
			t.Errorf("followers list lacks %q:\n%s", want, view)
		}
	}
	checkFrame(t, view, 80, 24)

	// hop to bob, the second follower
	m, _ = send(m, keyRunes("j"))
	m, msgs := send(m, enter)
	// alice's list stays under the spinner until bob arrives, and nothing
	// names alice as the profile being shown
	if trail, search := topRows(m); trail != "alice › bob" || search != "" {
		t.Errorf("while bob loads the top rows are %q and %q", trail, search)
	}
	if view := m.View(); !strings.Contains(view, "loading bob…") || !strings.Contains(view, "carol") || strings.Contains(bottomRow(m), "loading") {
		t.Errorf("bob's spinner is not over the dimmed list:\n%s", view)
	}
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	if m.state.User.Login != "bob" || m.state.CurrentView != state.UserView {
		t.Fatalf("enter on bob shows %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
	if strings.Contains(m.View(), "loading") {
		t.Errorf("spinner outlived the hop:\n%s", m.View())
	}

	// and on through bob's following, back to alice
	m = openList(t, m, state.FollowingTarget)
	m, msgs = send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	if m.state.User.Login != "alice" {
		t.Fatalf("hop from bob's following shows %q", m.state.User.Login)
	}

	if trail, _ := topRows(m); trail != "alice › bob › alice" {
		t.Errorf("trail after two hops is %q", trail)
	}

	// backspace retraces every step, shrinking the trail, then stops on
	// the first profile
	steps := []struct {
		login string
		view  state.View
		trail string
	}{
		{"bob", state.PeopleView, "alice › bob"},
		{"bob", state.UserView, "alice › bob"},
		{"alice", state.PeopleView, "alice"},
		{"alice", state.UserView, "alice"},
		{"alice", state.UserView, "alice"},
	}
	for i, want := range steps {
		m, _ = send(m, backspace)
		if m.state.User.Login != want.login || m.state.CurrentView != want.view {
			t.Fatalf("backspace %d shows %q in view %v, want %q in %v", i+1, m.state.User.Login, m.state.CurrentView, want.login, want.view)
		}
		if trail, _ := topRows(m); trail != want.trail {
			t.Errorf("backspace %d leaves the trail %q, want %q", i+1, trail, want.trail)
		}
		if want.view == state.PeopleView && want.login == "alice" {
			if p := m.state.People.SelectedPerson(); p == nil || p.Login != "bob" {
				t.Errorf("alice's list came back on %+v, want bob still selected", p)
			}
		}
	}
	if !strings.Contains(m.View(), "alpha") {
		t.Errorf("alice's profile did not come back:\n%s", m.View())
	}
	if n := strings.Count(strings.Join(client.calls, "\n"), "user alice"); n != 2 {
		t.Errorf("looked alice up %d times, want 2: going back must not refetch", n)
	}
}

func Test_BackspaceTypesInTheSearchBox(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"alice": {Login: "alice"}, "bob": {Login: "bob"}}}
	m := launch(t, client, "alice", 80, 24)
	m, res := lookup(t, m, "bob")
	m, _ = send(m, res)

	m, _ = send(m, keyRunes("/"))
	m, _ = send(m, keyRunes("carx"))
	m, _ = send(m, backspace)
	if m.state.User.Login != "bob" || m.state.Mode != state.InsertMode {
		t.Fatalf("backspace while typing went back to %q", m.state.User.Login)
	}
	if !strings.Contains(m.View(), "car") || strings.Contains(m.View(), "carx") {
		t.Errorf("backspace did not delete in the search box:\n%s", m.View())
	}
}

func Test_BackDropsAPendingHop(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	bob := find[commands.GetUserResponse](t, msgs)

	if trail, _ := topRows(m); trail != "alice › bob" {
		t.Fatalf("trail while bob loads is %q", trail)
	}

	m, _ = send(m, backspace) // before bob arrives
	m, _ = send(m, bob)
	if m.state.User.Login != "alice" || m.state.CurrentView != state.PeopleView {
		t.Errorf("a hop abandoned with backspace left the list: %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
	if m.loading.Visible() || strings.Contains(m.View(), "loading bob") {
		t.Errorf("spinner still showing with nothing pending:\n%s", m.View())
	}
	if trail, _ := topRows(m); trail != "alice" {
		t.Errorf("trail after cancelling the hop is %q, want alice", trail)
	}

	// a second backspace goes back a page as usual
	m, _ = send(m, backspace)
	if m.state.User.Login != "alice" || m.state.CurrentView != state.UserView {
		t.Errorf("second backspace left %q in view %v, want alice's profile", m.state.User.Login, m.state.CurrentView)
	}
}

func Test_ListPagesThroughLargeLists(t *testing.T) {
	client := &fakeClient{
		users:    map[string]models.User{"octocat": {Login: "octocat", FollowersCount: 120}},
		people:   map[string][]models.Person{"followers octocat": manyPeople("fan", 120)},
		pageSize: 50,
	}
	m := launch(t, client, "octocat", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	if len(m.state.People.List) != 50 {
		t.Fatalf("first page holds %d people, want 50", len(m.state.People.List))
	}

	// walk to the last follower; pages load as the selection nears the end
	for m.state.People.Selected < 119 {
		var msgs *inbox
		m, msgs = send(m, keyRunes("j"))
		if res, ok := firstOf[commands.GetPeopleResponse](msgs.all(t)); ok {
			m, _ = send(m, res)
		}
		checkFrame(t, m.View(), 80, 24)
		if p := m.state.People.SelectedPerson(); !strings.Contains(m.View(), "› "+p.Login) {
			t.Fatalf("selected %s is not on screen:\n%s", p.Login, m.View())
		}
		if m.state.People.Selected > 200 {
			t.Fatal("selection ran past the list")
		}
	}
	for _, call := range []string{"followers octocat after ", "followers octocat after 50", "followers octocat after 100"} {
		if !client.called(call) {
			t.Errorf("no request %q: calls %v", call, client.calls)
		}
	}

	// the end of the list stops there and asks for nothing more
	before := len(client.calls)
	m, _ = send(m, keyRunes("j"))
	if m.state.People.Selected != 119 || len(client.calls) != before {
		t.Errorf("j at the end moved to %d and made %d calls", m.state.People.Selected, len(client.calls)-before)
	}
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.state.People.Selected >= 119 {
		t.Errorf("ctrl+u did not page up")
	}
}

func Test_ListFailureAndEmptyList(t *testing.T) {
	client := &fakeClient{
		users:     map[string]models.User{"octocat": {Login: "octocat"}},
		peopleErr: errors.New("boom"),
	}
	m := launch(t, client, "octocat", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	if !strings.Contains(m.View(), "couldn't load followers") {
		t.Errorf("failed list does not say so:\n%s", m.View())
	}
	m, _ = send(m, enter) // nothing to open
	if m.state.CurrentView != state.PeopleView {
		t.Errorf("enter on an empty list left it")
	}

	client.peopleErr = nil
	m, _ = send(m, backspace)
	m = openList(t, m, state.FollowingTarget)
	if !strings.Contains(m.View(), "not following anyone yet") {
		t.Errorf("empty list does not say so:\n%s", m.View())
	}
}

func Test_TopReposAreLabelled(t *testing.T) {
	top := octocat()
	top.TopRepos = true
	client := &fakeClient{users: map[string]models.User{"octocat": top, "hubot": octocat()}}

	m := launch(t, client, "octocat", 80, 24)
	if !strings.Contains(m.View(), "top repositories") || !strings.Contains(m.View(), "alpha") {
		t.Errorf("top repos are not labelled as such:\n%s", m.View())
	}

	m, res := lookup(t, m, "hubot")
	m, _ = send(m, res)
	if strings.Contains(m.View(), "top repositories") {
		t.Errorf("pinned repos are labelled as top repos:\n%s", m.View())
	}
}

func Test_StarLandsOnAPageInHistory(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat(), "hubot": {Login: "hubot"}},
		starErr: errors.New("boom"),
	}
	m := launch(t, client, "octocat", 160, 50)
	m = focusRepo(t, m, "A")
	m, msgs := send(m, keyRunes("s"))
	failed := find[commands.StarResponse](t, msgs)

	m, res := lookup(t, m, "hubot")
	m, _ = send(m, res)
	m, _ = send(m, failed) // lands while octocat is in history
	m, _ = send(m, backspace)
	if r := m.state.Repo("A"); r == nil || r.ViewerHasStarred {
		t.Errorf("failed star still shows after going back: %+v", r)
	}
}
