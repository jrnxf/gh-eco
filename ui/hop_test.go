package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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
	if !spinnerOnTop(m) {
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
	if view := m.View(); !spinnerOnTop(m) || !strings.Contains(view, "carol") {
		t.Errorf("bob's spinner is not after his name over the dimmed list:\n%s", view)
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

func Test_BackDropsAPendingReadme(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}},
		people:  map[string][]models.Person{"followers alice": {{Login: "bob"}}},
		readmes: map[string]string{"bob/bob": "# bob readme"},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	// r opens the profile readme from the focused name
	for m.state.Focused().Kind != state.UserTarget {
		m, _ = send(m, keyRunes("j"))
	}
	m, msgs = send(m, keyRunes("r"))
	readme := find[commands.GetReadmeResponse](t, msgs)
	if !spinnerOnTop(m) {
		t.Fatalf("no readme spinner after r:\n%s", m.View())
	}

	m, _ = send(m, backspace) // before the readme arrives
	m, _ = send(m, readme)
	if m.state.User.Login != "bob" || m.state.CurrentView != state.UserView {
		t.Errorf("a readme abandoned with backspace left the profile: %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
	if m.loading.Visible() || strings.Contains(m.View(), "loading readme") {
		t.Errorf("spinner still showing with nothing pending:\n%s", m.View())
	}

	// a second backspace goes back a page as usual
	m, _ = send(m, backspace)
	if m.state.User.Login != "alice" || m.state.CurrentView != state.PeopleView {
		t.Errorf("second backspace left %q in view %v, want alice's list", m.state.User.Login, m.state.CurrentView)
	}
}

func Test_TrailDropsOnlyTheOldestThatDoNotFit(t *testing.T) {
	client := &fakeClient{
		users: map[string]models.User{
			"alice": {Login: "alice", FollowingCount: 1},
			"bob":   {Login: "bob", FollowingCount: 1},
			"carol": {Login: "carol", FollowingCount: 1},
			"dave":  {Login: "dave"},
		},
		people: map[string][]models.Person{
			"following alice": {{Login: "bob"}},
			"following bob":   {{Login: "carol"}},
			"following carol": {{Login: "dave"}},
		},
	}
	m := launch(t, client, "alice", 80, 24)
	for range 3 {
		m = openList(t, m, state.FollowingTarget)
		var msgs *inbox
		m, msgs = send(m, enter)
		m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	}
	if trail, _ := topRows(m); trail != "alice › bob › carol › dave" {
		t.Fatalf("trail after three hops is %q", trail)
	}

	full, cut := lipgloss.Width("alice › bob › carol › dave"), lipgloss.Width("… › bob › carol › dave")
	for _, tc := range []struct {
		width int
		want  string
	}{
		{full, "alice › bob › carol › dave"},
		{full - 1, "… › bob › carol › dave"},
		{cut - 1, "… › carol › dave"},
		{lipgloss.Width("… › dave"), "… › dave"},
	} {
		m.state.Layout.ContentWidth = tc.width + state.RightSlack
		if trail, _ := topRows(m); trail != tc.want {
			t.Errorf("trail in %d cells is %q, want %q", tc.width, trail, tc.want)
		}
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

func Test_ListShowsTheNextPageLoadingAtItsEnd(t *testing.T) {
	client := &fakeClient{
		users:    map[string]models.User{"octocat": {Login: "octocat", FollowersCount: 120}},
		people:   map[string][]models.Person{"followers octocat": manyPeople("fan", 120)},
		pageSize: 50,
	}
	m := launch(t, client, "octocat", 80, 24)
	m = openList(t, m, state.FollowersTarget)

	// walk to the last loaded follower, holding back the next page
	for m.state.People.Selected < 49 {
		m, _ = send(m, keyRunes("j"))
	}
	checkFrame(t, m.View(), 80, 24)
	if !spinnerShows(m, "loading followers…") {
		t.Errorf("no spinner under a full list while the next page loads:\n%s", m.View())
	}
	row49 := rowOf(m, "› fan-49")
	if row49 < 0 {
		t.Fatalf("selected fan-49 is not on screen:\n%s", m.View())
	}

	// k moves the cursor up a row; the list stays put
	m, _ = send(m, keyRunes("k"))
	if got := rowOf(m, "› fan-48"); got != row49-1 {
		t.Errorf("after k fan-48 is on row %d, want %d:\n%s", got, row49-1, m.View())
	}
	if !spinnerShows(m, "loading followers…") {
		t.Errorf("spinner gone after k while the next page loads:\n%s", m.View())
	}
	m, _ = send(m, keyRunes("j"))
	if got := rowOf(m, "› fan-49"); got != row49 {
		t.Errorf("after j fan-49 is on row %d, want %d:\n%s", got, row49, m.View())
	}
}

// rowOf is the frame row showing s, or -1.
func rowOf(m Model, s string) int {
	for i, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if strings.Contains(line, s) {
			return i
		}
	}
	return -1
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
	cat := octocat()
	cat.FollowersCount = 1
	client := &fakeClient{
		users:   map[string]models.User{"octocat": cat, "hubot": {Login: "hubot"}},
		people:  map[string][]models.Person{"followers octocat": {{Login: "hubot"}}},
		starErr: errors.New("boom"),
	}
	m := launch(t, client, "octocat", 160, 50)
	m = focusRepo(t, m, "A")
	m, msgs := send(m, keyRunes("s"))
	failed := find[commands.StarResponse](t, msgs)

	m = openList(t, m, state.FollowersTarget)
	m, msgs = send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	m, _ = send(m, failed) // lands while octocat is in history
	m, _ = send(m, backspace)
	if r := m.state.Repo("A"); r == nil || r.ViewerHasStarred {
		t.Errorf("failed star still shows after going back: %+v", r)
	}
}

var esc = tea.KeyMsg{Type: tea.KeyEsc}

// quits reports whether msg makes m quit.
func quits(m Model, msg tea.Msg) bool {
	_, cmd := m.Update(msg)
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func Test_EscLeavesASearchWithNothingLost(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))

	m, _ = send(m, keyRunes("/"))
	m, _ = send(m, keyRunes("car"))
	// the trail stays while a search is typed
	if trail, search := topRows(m); trail != "alice › bob" || !strings.Contains(search, "car") {
		t.Errorf("while typing the top rows are %q and %q", trail, search)
	}
	m, in := send(m, esc)
	if _, quit := firstOf[tea.QuitMsg](in.all(t)); quit {
		t.Fatal("esc in a search quit")
	}
	if m.state.Mode != state.NormalMode || m.state.User.Login != "bob" || m.state.CurrentView != state.UserView {
		t.Errorf("esc left mode %v on %q in view %v", m.state.Mode, m.state.User.Login, m.state.CurrentView)
	}
	if trail, search := topRows(m); trail != "alice › bob" || search != "" {
		t.Errorf("after esc the top rows are %q and %q", trail, search)
	}
	if len(client.calls) != 3 { // alice, her followers, bob
		t.Errorf("esc made calls: %v", client.calls)
	}

	// backspace still retraces the hop, and esc quits outside a search
	m, _ = send(m, backspace)
	if m.state.User.Login != "alice" || m.state.CurrentView != state.PeopleView {
		t.Errorf("backspace after esc shows %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
	// esc now closes layers like backspace, and quits at the top level
	if quits(m, esc) {
		t.Error("esc quit with a page to go back to")
	}
	if m.state.User.Login != "alice" || m.state.CurrentView != state.UserView {
		t.Errorf("esc shows %q in view %v, want alice's profile", m.state.User.Login, m.state.CurrentView)
	}
	if !quits(m, esc) {
		t.Error("esc at the top level does not quit")
	}
}

func Test_EscQuitsTheSearchBoxAtLaunch(t *testing.T) {
	m := newModel(t, &fakeClient{})
	if !quits(m, esc) {
		t.Error("esc in the launch search box does not quit")
	}
}

func Test_SearchStartsANewTrail(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}, "carol": {Login: "carol"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))

	m, carol := lookup(t, m, "carol")
	// the trail shows where backspace goes once carol arrives: nowhere
	if trail, _ := topRows(m); trail != "carol" {
		t.Errorf("trail while carol loads is %q", trail)
	}
	m, _ = send(m, carol)
	if trail, _ := topRows(m); trail != "carol" || len(m.state.History) != 0 {
		t.Errorf("after the search the trail is %q with %d pages behind", trail, len(m.state.History))
	}
	m, _ = send(m, backspace)
	if m.state.User.Login != "carol" || m.state.CurrentView != state.UserView {
		t.Errorf("backspace from a searched profile shows %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
}

func Test_FailedSearchKeepsTheTrail(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))

	m, res := lookup(t, m, "nobody")
	m, _ = send(m, res)
	if trail, _ := topRows(m); trail != "alice › bob" {
		t.Errorf("trail after a failed search is %q", trail)
	}
	// backspace still reaches every page from before the search
	m, _ = send(m, backspace)
	if m.state.User.Login != "bob" {
		t.Errorf("backspace after a failed search shows %q", m.state.User.Login)
	}
}

func Test_BackspaceCancelsASearchKeepingTheTrail(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}, "carol": {Login: "carol"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))

	m, carol := lookup(t, m, "carol")
	m, _ = send(m, backspace) // before carol arrives
	m, _ = send(m, carol)
	if trail, _ := topRows(m); m.state.User.Login != "bob" || trail != "alice › bob" {
		t.Errorf("a cancelled search shows %q with trail %q", m.state.User.Login, trail)
	}
}

// contentRows is the frame under its top two rows and above the help bar.
func contentRows(m Model) []string {
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	return lines[2 : len(lines)-1]
}

func Test_SearchNeverMovesTheContent(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}, "bob": {Login: "bob"}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	before := contentRows(m)
	frame := len(strings.Split(m.View(), "\n"))

	m, _ = send(m, keyRunes("/"))
	if got := contentRows(m); strings.Join(got, "\n") != strings.Join(before, "\n") || len(strings.Split(m.View(), "\n")) != frame {
		t.Errorf("/ moved the content:\n%s", m.View())
	}
	m, _ = send(m, esc)
	if got := contentRows(m); strings.Join(got, "\n") != strings.Join(before, "\n") {
		t.Errorf("esc moved the content:\n%s", m.View())
	}

	// the same on the list, and with the search already submitted
	m = openList(t, m, state.FollowersTarget)
	listRows := rowOf(m, "bob")
	m, _ = send(m, keyRunes("/"))
	if got := rowOf(m, "bob"); got != listRows {
		t.Errorf("/ moved the list from row %d to %d", listRows, got)
	}
}

func Test_ListFirstPageSpinnerIsOnTheTopRow(t *testing.T) {
	client := &fakeClient{
		users:  map[string]models.User{"alice": {Login: "alice", FollowersCount: 1}},
		people: map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	for m.state.Focused().Kind != state.FollowersTarget {
		m, _ = send(m, keyRunes("j"))
	}
	m, _ = send(m, enter)
	if !spinnerOnTop(m) {
		t.Errorf("no spinner on the top row while the first page loads:\n%s", m.View())
	}
}

func Test_TypedSearchKeepsTheTrailOnOneNarrowLine(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{
		"alice": {Login: "alice", FollowersCount: 1},
	}}
	m := launch(t, client, "alice", 80, 24)
	// a long hop trail
	for _, l := range []string{"bob-the-builder", "carol-of-the-bells", "dave-from-accounting", "erin-brockovich"} {
		m.state.History = append(m.state.History, state.Page{User: models.User{Login: l}})
	}
	m, _ = send(m, keyRunes("/"))
	m, _ = send(m, keyRunes("some-long-search-term"))
	checkFrame(t, m.View(), 80, 24)
	row := topRow(m)
	if !strings.HasPrefix(row, "> some-long-search-term") || !strings.Contains(row, "…") || !strings.HasSuffix(row, "alice") {
		t.Errorf("top row is not the search with a left-truncated trail: %q", row)
	}
}

func Test_LongSearchStaysOnTheTopLine(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"alice": {Login: "alice"}}}
	m := launch(t, client, "alice", 40, 24)
	before := contentRows(m)

	m, _ = send(m, keyRunes("/"))
	m, _ = send(m, keyRunes("a-very-long-login-name-of-39-characters"))
	checkFrame(t, m.View(), 40, 24)
	if got := contentRows(m); strings.Join(got, "\n") != strings.Join(before, "\n") {
		t.Errorf("a long search moved the content:\n%s", m.View())
	}
	if row := topRow(m); !strings.HasSuffix(row, "characters") {
		t.Errorf("a long search does not keep its end in sight: %q", row)
	}

	// with the cursor moved back into the search
	for range 30 {
		m, _ = send(m, tea.KeyMsg{Type: tea.KeyLeft})
	}
	checkFrame(t, m.View(), 40, 24)
	if got := contentRows(m); strings.Join(got, "\n") != strings.Join(before, "\n") {
		t.Errorf("moving the cursor in a long search moved the content:\n%s", m.View())
	}
}

// profileWithRepo is a profile whose repo has a readme.
func profileWithRepo(login string) models.User {
	return models.User{
		Login: login, Url: "https://github.com/" + login, FollowersCount: 1,
		Repos: []models.Repo{{Id: login + "-repo", Name: "alpha", Owner: struct{ Login string }{login}}},
	}
}

func Test_BackspaceInAReadmeClosesOnlyTheReadme(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"alice": profileWithRepo("alice"), "bob": profileWithRepo("bob")},
		readmes: map[string]string{"bob/alpha": "# bob readme"},
		people:  map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	m = focusRepo(t, m, "bob-repo")
	m, msgs = send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if m.state.CurrentView != state.ReadmeView {
		t.Fatalf("readme did not open")
	}
	history := len(m.state.History)

	m, _ = send(m, backspace)
	if m.state.CurrentView != state.UserView || m.state.User.Login != "bob" || len(m.state.History) != history {
		t.Errorf("backspace in a readme left view %v on %q with %d pages of history, want bob's profile with %d",
			m.state.CurrentView, m.state.User.Login, len(m.state.History), history)
	}
	if trail, _ := topRows(m); trail != "alice › bob" {
		t.Errorf("trail after closing the readme is %q", trail)
	}
	// the next backspace is the page back
	m, _ = send(m, backspace)
	if m.state.CurrentView != state.PeopleView {
		t.Errorf("second backspace shows view %v, want the list", m.state.CurrentView)
	}
}

var qKey = keyRunes("q")

// press delivers msg and reports whether it quit the program. Unlike
// quits, it keeps the model the key produced.
func press(m Model, msg tea.Msg) (Model, bool) {
	next, cmd := m.Update(msg)
	if cmd == nil {
		return next.(Model), false
	}
	_, quit := cmd().(tea.QuitMsg)
	return next.(Model), quit
}

// hopFixture is alice, who has a repo and one follower, bob.
func hopFixture() *fakeClient {
	return &fakeClient{
		users:   map[string]models.User{"alice": profileWithRepo("alice"), "bob": profileWithRepo("bob")},
		readmes: map[string]string{"alice/alpha": "# alice readme", "bob/alpha": "# bob readme"},
		people:  map[string][]models.Person{"followers alice": {{Login: "bob"}}},
	}
}

func Test_LatestKeyWinsOverAPendingHop(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  tea.KeyMsg
	}{{"j", keyRunes("j")}, {"k", keyRunes("k")}, {"enter", enter}} {
		t.Run(tc.name, func(t *testing.T) {
			m := launch(t, hopFixture(), "alice", 80, 24)
			m, bob := lookup(t, m, "bob")
			// the search put focus back on the profile; park it on the
			// followers count without a key press
			for m.state.Focused().Kind != state.FollowersTarget {
				m.state.FocusNext()
			}

			m, in := send(m, tc.key)
			if m.pendingLogin != "" || m.loading.Visible() && tc.key.Type != tea.KeyEnter {
				t.Errorf("%s left the hop pending: %q", tc.name, m.pendingLogin)
			}
			if trail, _ := topRows(m); !strings.HasPrefix(trail, "alice") || strings.Contains(trail, "bob") {
				t.Errorf("trail after %s is %q, want alice alone", tc.name, trail)
			}
			if tc.key.Type == tea.KeyEnter {
				// the key acted at once: alice's followers opened
				if m.state.CurrentView != state.PeopleView {
					t.Errorf("enter did not open the list: view %v", m.state.CurrentView)
				}
				m, _ = send(m, find[commands.GetPeopleResponse](t, in))
			}
			// bob's late answer changes nothing
			before := m.state.CurrentView
			m, _ = send(m, bob)
			if m.state.User.Login != "alice" || m.state.CurrentView != before || len(m.state.History) > 1 {
				t.Errorf("a dropped hop landed: %q in view %v", m.state.User.Login, m.state.CurrentView)
			}
		})
	}
}

func Test_EscClosesTheTopLayer(t *testing.T) {
	// readme -> its profile, history untouched
	m := launch(t, hopFixture(), "alice", 80, 24)
	m = focusRepo(t, m, "alice-repo")
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	m, quit := press(m, esc)
	if quit || m.state.CurrentView != state.UserView || m.state.User.Login != "alice" {
		t.Errorf("esc in a readme: view %v on %q", m.state.CurrentView, m.state.User.Login)
	}

	// pending readme -> dropped, profile stays
	m, _ = send(m, keyRunes("r"))
	if !m.readmePending {
		t.Fatal("readme not pending")
	}
	if m, quit = press(m, esc); quit || m.readmePending || m.loading.Visible() {
		t.Errorf("esc did not drop the pending readme")
	}

	// list -> profile
	m = openList(t, m, state.FollowersTarget)
	if m, quit = press(m, esc); quit || m.state.CurrentView != state.UserView || len(m.state.History) != 0 {
		t.Errorf("esc in a list: view %v, %d pages of history", m.state.CurrentView, len(m.state.History))
	}

	// pending hop -> dropped; the hopped profile -> the page it came from
	m, res := lookup(t, m, "bob")
	if m, quit = press(m, esc); quit || m.pendingLogin != "" || m.state.User.Login != "alice" {
		t.Errorf("esc did not drop the pending hop")
	}
	m, _ = send(m, res)
	if m.state.User.Login != "alice" {
		t.Errorf("a dropped hop landed")
	}
	m = openList(t, m, state.FollowersTarget)
	m, msgs = send(m, enter)
	m, _ = send(m, find[commands.GetUserResponse](t, msgs))
	if m, quit = press(m, esc); quit || m.state.CurrentView != state.PeopleView {
		t.Errorf("esc on a hopped profile shows view %v", m.state.CurrentView)
	}

	// the top level quits
	m, _ = press(m, esc)
	if _, quit = press(m, esc); !quit {
		t.Errorf("esc at the top level does not quit (user %q, history %d)", m.state.User.Login, len(m.state.History))
	}
}

func Test_QAlwaysQuitsExceptWhileTyping(t *testing.T) {
	m := launch(t, hopFixture(), "alice", 80, 24)
	if _, quit := press(m, qKey); !quit {
		t.Error("q on a profile does not quit")
	}
	m = openList(t, m, state.FollowersTarget)
	if _, quit := press(m, qKey); !quit {
		t.Error("q in a list does not quit")
	}
	m, _ = send(m, esc)
	m = focusRepo(t, m, "alice-repo")
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if _, quit := press(m, qKey); !quit {
		t.Error("q in a readme does not quit")
	}
	m, _ = send(m, esc)
	m, _ = lookup(t, m, "bob")
	if _, quit := press(m, qKey); !quit {
		t.Error("q over a pending hop does not quit")
	}

	// in the search box q is a letter
	m, _ = send(m, keyRunes("/"))
	m, quit := press(m, qKey)
	if quit {
		t.Fatal("q quit from the search box")
	}
	if _, search := topRows(m); !strings.Contains(search, "q") {
		t.Errorf("q was not typed: %q", search)
	}
}

func Test_SlashOpensSearchInsideAReadme(t *testing.T) {
	m := launch(t, hopFixture(), "alice", 80, 24)
	m = focusRepo(t, m, "alice-repo")
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))

	m, _ = send(m, keyRunes("/"))
	if m.state.Mode != state.InsertMode || m.state.CurrentView != state.UserView {
		t.Fatalf("/ in a readme left mode %v in view %v", m.state.Mode, m.state.CurrentView)
	}
	m, _ = send(m, keyRunes("bo"))
	if _, search := topRows(m); !strings.Contains(search, "bo") {
		t.Errorf("typing in the search after / in a readme: %q", search)
	}
	checkFrame(t, m.View(), 80, 24)
	// esc leaves the search on the profile
	m, _ = send(m, esc)
	if m.state.Mode != state.NormalMode || m.state.User.Login != "alice" {
		t.Errorf("esc after / in a readme: mode %v on %q", m.state.Mode, m.state.User.Login)
	}
}

func Test_UnboundKeysKeepAPendingLoad(t *testing.T) {
	unbound := []tea.KeyMsg{keyRunes("x"), {Type: tea.KeyTab}, {Type: tea.KeyShiftTab}}

	// a pending hop
	m := launch(t, hopFixture(), "alice", 80, 24)
	m, bob := lookup(t, m, "bob")
	for _, k := range unbound {
		m, _ = send(m, k)
	}
	if m.pendingLogin != "bob" || !spinnerOnTop(m) {
		t.Fatalf("an unbound key dropped the pending hop:\n%s", m.View())
	}
	m, _ = send(m, bob)
	if m.state.User.Login != "bob" {
		t.Errorf("the kept hop did not land: %q", m.state.User.Login)
	}

	// a pending readme
	m = focusRepo(t, m, "bob-repo")
	m, msgs := send(m, keyRunes("r"))
	readme := find[commands.GetReadmeResponse](t, msgs)
	for _, k := range unbound {
		m, _ = send(m, k)
	}
	if !m.readmePending || !spinnerOnTop(m) {
		t.Fatalf("an unbound key dropped the pending readme:\n%s", m.View())
	}
	m, _ = send(m, readme)
	if m.state.CurrentView != state.ReadmeView {
		t.Errorf("the kept readme did not open: view %v", m.state.CurrentView)
	}
}

func Test_InertKeysKeepAPendingHopFromAList(t *testing.T) {
	client := hopFixture()
	client.people["followers alice"] = []models.Person{{Login: "bob"}, {Login: "carol"}}
	m := launch(t, client, "alice", 80, 24)
	m = openList(t, m, state.FollowersTarget)
	m, msgs := send(m, enter)
	bob := find[commands.GetUserResponse](t, msgs)

	// a list has no readme or repo, and k is already at the top
	for _, k := range []tea.KeyMsg{keyRunes("x"), keyRunes("r"), keyRunes("s"), keyRunes("k")} {
		m, _ = send(m, k)
		if m.pendingLogin != "bob" || !spinnerOnTop(m) {
			t.Fatalf("%q did nothing yet dropped the pending hop:\n%s", k.String(), m.View())
		}
	}

	// j moves the selection, so it wins over the hop
	m, _ = send(m, keyRunes("j"))
	if m.pendingLogin != "" || m.loading.Visible() {
		t.Errorf("j left the hop pending: %q", m.pendingLogin)
	}
	if p := m.state.People.SelectedPerson(); p == nil || p.Login != "carol" {
		t.Errorf("j did not move the selection: %+v", p)
	}
	m, _ = send(m, bob)
	if m.state.User.Login != "alice" || m.state.CurrentView != state.PeopleView {
		t.Errorf("a dropped hop landed: %q in view %v", m.state.User.Login, m.state.CurrentView)
	}
}

func Test_HelpOffersBackWheneverBackWorks(t *testing.T) {
	// launched with a login: the first lookup can be dropped
	client := hopFixture()
	m, _ := send(New(client, "alice"), tea.WindowSizeMsg{Width: 200, Height: 24})
	if !strings.Contains(ansi.Strip(bottomRow(m)), "back") {
		t.Errorf("help while the first profile loads has no back: %q", ansi.Strip(bottomRow(m)))
	}

	// a first profile with no history and a readme pending
	m = launch(t, client, "alice", 200, 24)
	if strings.Contains(ansi.Strip(bottomRow(m)), "back") {
		t.Errorf("help offers back with nothing to go back to: %q", ansi.Strip(bottomRow(m)))
	}
	m = focusRepo(t, m, "alice-repo")
	m, msgs := send(m, keyRunes("r"))
	if !strings.Contains(ansi.Strip(bottomRow(m)), "back") {
		t.Errorf("help while a readme loads has no back: %q", ansi.Strip(bottomRow(m)))
	}

	// / works in a readme, and the help says so
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if m.state.CurrentView != state.ReadmeView {
		t.Fatalf("readme did not open")
	}
	if !strings.Contains(ansi.Strip(bottomRow(m)), "/ focus input") {
		t.Errorf("readme help does not offer /: %q", ansi.Strip(bottomRow(m)))
	}
}
