package ui

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

// fakeClient serves canned users and readmes, caches readmes like the
// real client and records every call.
type fakeClient struct {
	mu      sync.Mutex
	calls   []string
	users   map[string]models.User
	readmes map[string]string
	cache   map[string]string
	// starErr fails every star and unstar.
	starErr error
	// people maps "followers login" or "following login" to every entry;
	// pageSize splits it into pages, all of it in one by default.
	people   map[string][]models.Person
	pageSize int
	// peopleErr fails every list page.
	peopleErr error
}

func (f *fakeClient) record(call string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
}

func (f *fakeClient) called(call string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Contains(f.calls, call)
}

func (f *fakeClient) User(_ context.Context, login string) (models.User, error) {
	f.record("user " + login)
	if u, ok := f.users[login]; ok {
		return u, nil
	}
	return models.User{}, github.ErrNotFound
}

func (f *fakeClient) People(_ context.Context, login string, kind models.PeopleKind, after string) (models.PeoplePage, error) {
	list := "followers " + login
	if kind == models.Following {
		list = "following " + login
	}
	f.record(list + " after " + after)
	if f.peopleErr != nil {
		return models.PeoplePage{}, f.peopleErr
	}
	all := f.people[list]
	start := 0
	if after != "" {
		start, _ = strconv.Atoi(after)
	}
	end := len(all)
	if f.pageSize > 0 {
		end = min(end, start+f.pageSize)
	}
	page := models.PeoplePage{People: all[start:end], TotalCount: len(all), HasNextPage: end < len(all)}
	if page.HasNextPage {
		page.EndCursor = strconv.Itoa(end)
	}
	return page, nil
}

func (f *fakeClient) Readme(_ context.Context, owner, name string) (string, error) {
	f.record("readme " + owner + "/" + name)
	r, ok := f.readmes[owner+"/"+name]
	if !ok {
		return "", github.ErrNotFound
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.cache == nil {
		f.cache = map[string]string{}
	}
	f.cache[owner+"/"+name] = r
	return r, nil
}

func (f *fakeClient) OrgReadme(_ context.Context, org string) (string, error) {
	f.record("org readme " + org)
	return "# " + org, nil
}

func (f *fakeClient) CachedReadme(owner, name string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.cache[owner+"/"+name]
	return r, ok
}

func (f *fakeClient) CachedOrgReadme(string) (string, bool) {
	return "", false
}

func (f *fakeClient) Star(_ context.Context, id string) (models.Starrable, error) {
	f.record("star " + id)
	if f.starErr != nil {
		return models.Starrable{}, f.starErr
	}
	return models.Starrable{Id: id, ViewerHasStarred: true, StargazerCount: 1}, nil
}

func (f *fakeClient) Unstar(_ context.Context, id string) (models.Starrable, error) {
	f.record("unstar " + id)
	if f.starErr != nil {
		return models.Starrable{}, f.starErr
	}
	return models.Starrable{Id: id}, nil
}

func octocat() models.User {
	return models.User{
		Login: "octocat",
		Url:   "https://github.com/octocat",
		Repos: []models.Repo{
			{Id: "A", Name: "alpha", Owner: struct{ Login string }{"octocat"}},
			{Id: "B", Name: "beta", Owner: struct{ Login string }{"octocat"}},
		},
	}
}

// inbox collects the messages a command and every command it batches
// produce. Each command runs in its own goroutine, so a slow one (a cursor
// blink, a spinner or timer tick) never holds up the others.
type inbox struct {
	ch      chan tea.Msg
	running sync.WaitGroup
	got     []tea.Msg
}

func run(cmd tea.Cmd) *inbox {
	in := &inbox{ch: make(chan tea.Msg, 256)}
	in.launch(cmd)
	return in
}

func (in *inbox) launch(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	in.running.Add(1)
	go func() {
		defer in.running.Done()
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				in.launch(c)
			}
			return
		}
		if msg != nil {
			in.ch <- msg
		}
	}()
}

// waitLimit only bounds a broken test; a passing one never waits for it.
const waitLimit = 5 * time.Second

// find returns the first message of type T, waiting until one arrives.
func find[T tea.Msg](t *testing.T, in *inbox) T {
	t.Helper()
	for _, msg := range in.got {
		if m, ok := msg.(T); ok {
			return m
		}
	}
	timeout := time.After(waitLimit)
	for {
		select {
		case msg := <-in.ch:
			in.got = append(in.got, msg)
			if m, ok := msg.(T); ok {
				return m
			}
		case <-timeout:
			var zero T
			t.Fatalf("no %T after %v; got %d other messages", zero, waitLimit, len(in.got))
			return zero
		}
	}
}

// all waits for every command to finish and returns all their messages.
// Use it only where no command sleeps, such as a blink or a tick.
func (in *inbox) all(t *testing.T) []tea.Msg {
	t.Helper()
	done := make(chan struct{})
	go func() { in.running.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(waitLimit):
		t.Fatalf("commands still running after %v", waitLimit)
	}
	for {
		select {
		case msg := <-in.ch:
			in.got = append(in.got, msg)
		default:
			return in.got
		}
	}
}

// firstOf returns the first message of type T, if any.
func firstOf[T tea.Msg](msgs []tea.Msg) (T, bool) {
	for _, msg := range msgs {
		if m, ok := msg.(T); ok {
			return m, true
		}
	}
	var zero T
	return zero, false
}

// send feeds msg to m and returns the updated model and the messages its
// commands produce.
func send(m Model, msg tea.Msg) (Model, *inbox) {
	next, cmd := m.Update(msg)
	return next.(Model), run(cmd)
}

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

var enter = tea.KeyMsg{Type: tea.KeyEnter}

// lookup types login, submits it and returns the pending lookup's response
// without delivering it.
func lookup(t *testing.T, m Model, login string) (Model, commands.GetUserResponse) {
	t.Helper()
	if m.state.Mode == state.NormalMode {
		m, _ = send(m, keyRunes("/"))
	}
	m, _ = send(m, keyRunes(login))
	m, msgs := send(m, enter)
	m, msgs = send(m, find[commands.SearchSubmitted](t, msgs))
	return m, find[commands.GetUserResponse](t, msgs)
}

// focusRepo presses j until the repo with the given id has focus.
func focusRepo(t *testing.T, m Model, id string) Model {
	t.Helper()
	for range m.state.Targets() {
		if r := m.state.FocusedRepo(); r != nil && r.Id == id {
			return m
		}
		m, _ = send(m, keyRunes("j"))
	}
	t.Fatalf("no repo %s to focus on %q", id, m.state.User.Login)
	return m
}

func newModel(t *testing.T, client *fakeClient) Model {
	t.Helper()
	m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: 160, Height: 50})
	return m
}

func Test_StarResponseLandsAfterFocusMoved(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)

	m = focusRepo(t, m, "A")
	m, msgs := send(m, keyRunes("s"))
	starred := find[commands.StarResponse](t, msgs)
	if !client.called("star A") {
		t.Fatalf("s on alpha did not star it: calls %v", client.calls)
	}

	// move to beta before alpha's star lands
	m = focusRepo(t, m, "B")
	m, _ = send(m, starred)

	if r := m.state.FocusedRepo(); r == nil || r.Id != "B" || r.ViewerHasStarred {
		t.Fatalf("focused repo = %+v, want unstarred beta", r)
	}
	if !m.state.User.Repos[0].ViewerHasStarred {
		t.Errorf("alpha lost its star")
	}

	_, msgs = send(m, keyRunes("s"))
	find[commands.StarResponse](t, msgs)
	if !client.called("star B") || client.called("unstar B") {
		t.Errorf("s on beta must star it: calls %v", client.calls)
	}
}

func Test_SupersededSearchIsDropped(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{
		"alice": {Login: "alice"},
		"bob":   {Login: "bob"},
	}}
	m := newModel(t, client)

	m, alice := lookup(t, m, "alice")
	m, _ = send(m, keyRunes("/")) // search again before alice arrives
	m, bob := lookup(t, m, "bob")

	m, _ = send(m, bob)
	m, _ = send(m, alice) // the slower, older lookup lands last

	if got := m.state.User.Login; got != "bob" {
		t.Errorf("showing %q, want bob, the latest search", got)
	}
}

func Test_ReadmeToggle(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat()},
		readmes: map[string]string{"octocat/beta": "# beta readme"},
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "B")

	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if m.state.CurrentView != state.ReadmeView {
		t.Fatalf("view = %v after r, want the readme", m.state.CurrentView)
	}
	if !client.called("readme octocat/beta") {
		t.Errorf("fetched the wrong readme: calls %v", client.calls)
	}

	m, _ = send(m, keyRunes("r"))
	if m.state.CurrentView != state.UserView {
		t.Fatalf("view = %v after second r, want the user", m.state.CurrentView)
	}
	if r := m.state.FocusedRepo(); r == nil || r.Id != "B" {
		t.Errorf("focus moved to %+v while reading", r)
	}
}

func Test_StaleReadmeIsDropped(t *testing.T) {
	newClient := func() *fakeClient {
		return &fakeClient{
			users: map[string]models.User{"octocat": octocat(), "bob": {Login: "bob"}},
			readmes: map[string]string{
				"octocat/alpha": "# alpha readme",
				"octocat/beta":  "# beta readme",
			},
		}
	}
	// focusAlpha loads octocat and focuses alpha.
	focusAlpha := func(t *testing.T, client *fakeClient) Model {
		m := newModel(t, client)
		m, res := lookup(t, m, "octocat")
		m, _ = send(m, res)
		m = focusRepo(t, m, "A")
		return m
	}

	t.Run("typing a new search", func(t *testing.T) {
		m := focusAlpha(t, newClient())
		m, msgs := send(m, keyRunes("r"))
		readme := find[commands.GetReadmeResponse](t, msgs)

		m, _ = send(m, keyRunes("/"))
		m, _ = send(m, keyRunes("bo"))
		m, _ = send(m, readme)

		if m.state.CurrentView != state.UserView || m.state.Mode != state.InsertMode {
			t.Fatalf("view = %v, mode = %v, want the search box still typing", m.state.CurrentView, m.state.Mode)
		}
		if !strings.Contains(m.View(), "bo") {
			t.Errorf("search box is hidden:\n%s", m.View())
		}
	})

	t.Run("after a new search lands", func(t *testing.T) {
		m := focusAlpha(t, newClient())
		m, msgs := send(m, keyRunes("r"))
		readme := find[commands.GetReadmeResponse](t, msgs)

		m, _ = send(m, keyRunes("/"))
		m, res := lookup(t, m, "bob")
		m, _ = send(m, res)
		m, _ = send(m, readme)

		if m.state.CurrentView != state.UserView || m.state.User.Login != "bob" {
			t.Errorf("octocat's readme opened over bob: view = %v", m.state.CurrentView)
		}
	})

	t.Run("superseded by a newer readme", func(t *testing.T) {
		m := focusAlpha(t, newClient())
		m, msgs := send(m, keyRunes("r"))
		alpha := find[commands.GetReadmeResponse](t, msgs)
		m, _ = send(m, keyRunes("j")) // beta
		m, msgs = send(m, keyRunes("r"))
		beta := find[commands.GetReadmeResponse](t, msgs)

		m, _ = send(m, alpha) // the older request lands first
		if m.state.CurrentView != state.UserView {
			t.Fatalf("alpha's readme opened after r on beta")
		}
		m, _ = send(m, beta)
		if m.state.CurrentView != state.ReadmeView || !strings.Contains(m.View(), "beta readme") {
			t.Errorf("beta's readme did not open:\n%s", m.View())
		}
	})
}

func Test_ReadmeOpenedDuringSearch(t *testing.T) {
	// start loads octocat, submits hubot and presses r on octocat's alpha
	// card before hubot arrives. It returns both pending responses.
	start := func(t *testing.T) (Model, commands.GetUserResponse, commands.GetReadmeResponse) {
		client := &fakeClient{
			users:   map[string]models.User{"octocat": octocat(), "hubot": {Login: "hubot"}},
			readmes: map[string]string{"octocat/alpha": "# alpha readme"},
		}
		m := newModel(t, client)
		m, res := lookup(t, m, "octocat")
		m, _ = send(m, res)
		m, _ = send(m, keyRunes("/"))
		m, hubot := lookup(t, m, "hubot")
		m = focusRepo(t, m, "A")
		m, msgs := send(m, keyRunes("r"))
		return m, hubot, find[commands.GetReadmeResponse](t, msgs)
	}

	t.Run("readme lands first", func(t *testing.T) {
		m, hubot, readme := start(t)
		m, _ = send(m, readme)
		if !m.loading.Visible() {
			t.Errorf("spinner stopped while hubot's search is still out")
		}
		m, _ = send(m, hubot)
		if m.state.CurrentView != state.UserView || m.state.User.Login != "hubot" {
			t.Errorf("view = %v for %q, want hubot's profile", m.state.CurrentView, m.state.User.Login)
		}
		if m.loading.Visible() {
			t.Errorf("spinner still showing with nothing pending")
		}
	})

	t.Run("search lands first", func(t *testing.T) {
		m, hubot, readme := start(t)
		m, _ = send(m, hubot)
		m, _ = send(m, readme)
		if m.state.CurrentView != state.UserView || m.state.User.Login != "hubot" {
			t.Errorf("octocat's readme opened over hubot: view = %v", m.state.CurrentView)
		}
		if m.loading.Visible() {
			t.Errorf("spinner still showing with nothing pending")
		}
	})
}

func Test_FailedSearchClearsPreviousUser(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m, _ = send(m, keyRunes("j"))

	m, _ = send(m, keyRunes("/"))
	m, res = lookup(t, m, "ghost")
	m, _ = send(m, res)

	if m.state.User.Login != "" || m.state.Focused().Kind != state.NoTarget {
		t.Errorf("octocat is still loaded after a failed search: %+v", m.state.User)
	}
	view := m.View()
	if !strings.Contains(view, `no user or organization named "ghost"`) {
		t.Errorf("view does not say ghost was not found:\n%s", view)
	}
	if strings.Contains(view, "toggle readme") {
		t.Errorf("help bar still offers actions on the previous user")
	}

	// keys meant for the old user do nothing
	_, msgs := send(m, keyRunes("r"))
	if _, ok := firstOf[commands.GetReadmeResponse](msgs.all(t)); ok || client.called("readme octocat/octocat") {
		t.Errorf("r after a failed search fetched %v", client.calls)
	}
}

func Test_OrganizationProfile(t *testing.T) {
	acme := models.User{
		IsOrganization: true,
		Login:          "acme",
		Name:           "Acme",
		Bio:            "we make things",
		Repos:          []models.Repo{{Id: "R_9", Name: "anvil", Owner: struct{ Login string }{"acme"}}},
	}
	client := &fakeClient{users: map[string]models.User{"acme": acme}}
	m := newModel(t, client)
	m, res := lookup(t, m, "acme")
	m, _ = send(m, res)

	view := m.View()
	for _, want := range []string{"Acme", "we make things", "anvil"} {
		if !strings.Contains(view, want) {
			t.Errorf("view lacks %q", want)
		}
	}
	for _, unwanted := range []string{"followers", "contributions"} {
		if strings.Contains(view, unwanted) {
			t.Errorf("organization view shows %q", unwanted)
		}
	}

	m, _ = send(m, keyRunes("j"))
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if !client.called("org readme acme") || m.state.CurrentView != state.ReadmeView {
		t.Errorf("r on an organization must open its profile readme: calls %v", client.calls)
	}
}

func Test_NoticeKeepsTheView(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat()},
		readmes: map[string]string{"octocat/alpha": "# alpha readme"},
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "A")
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))

	// a failed star shows a notice over the help bar, not instead of the page
	m, msgs = send(m, commands.StarResponse{Err: errors.New("boom")})
	m, _ = send(m, find[commands.SetMessage](t, msgs))
	lines := strings.Split(m.View(), "\n")
	if len(lines) != 50 || !strings.Contains(lines[49], "star failed") {
		t.Fatalf("notice is not on the bottom row of a full frame: %d lines, bottom %q", len(lines), lines[len(lines)-1])
	}
	if m.state.CurrentView != state.ReadmeView || !strings.Contains(m.View(), "alpha readme") {
		t.Errorf("the readme is hidden while the notice shows")
	}

	// keys keep working while it shows
	m, _ = send(m, keyRunes("r"))
	if m.state.CurrentView != state.UserView {
		t.Errorf("r was swallowed by the notice")
	}
}

func Test_StarGhEcoIsInstantAndRevertsOnFailure(t *testing.T) {
	// a profile that pins gh-eco, so the star shows on its card
	owner := models.User{Login: "jrnxf", Repos: []models.Repo{
		{Id: github.GhEcoRepoID, Name: "gh-eco", StarsCount: 482},
	}}
	client := &fakeClient{
		users:   map[string]models.User{"jrnxf": owner},
		starErr: errors.New("You do not have permission to star this repository"),
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "jrnxf")
	m, _ = send(m, res)

	m, msgs := send(m, keyRunes("!"))
	if r := m.state.Repo(github.GhEcoRepoID); !r.ViewerHasStarred || r.StarsCount != 483 {
		t.Fatalf("star not shown at once: %+v", r)
	}
	if msg := find[commands.SetMessage](t, msgs); !strings.Contains(msg.Content, "tysm") {
		t.Errorf("no thanks right away: %q", msg.Content)
	}

	m, msgs = send(m, find[commands.StarResponse](t, msgs))
	if r := m.state.Repo(github.GhEcoRepoID); r.ViewerHasStarred || r.StarsCount != 482 {
		t.Errorf("failed star not reverted: %+v", r)
	}
	m, _ = send(m, find[commands.SetMessage](t, msgs))
	if !strings.Contains(bottomRow(m), "star failed: You do not have permission") {
		t.Errorf("notice does not say why: %q", bottomRow(m))
	}
}

func Test_StarGhEcoKeepsAnExistingStarOnFailure(t *testing.T) {
	owner := models.User{Login: "jrnxf", Repos: []models.Repo{
		{Id: github.GhEcoRepoID, Name: "gh-eco", StarsCount: 482, ViewerHasStarred: true},
	}}
	client := &fakeClient{
		users:   map[string]models.User{"jrnxf": owner},
		starErr: errors.New("boom"),
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "jrnxf")
	m, _ = send(m, res)

	m, msgs := send(m, keyRunes("!"))
	m, _ = send(m, find[commands.StarResponse](t, msgs))
	if r := m.state.Repo(github.GhEcoRepoID); !r.ViewerHasStarred || r.StarsCount != 482 {
		t.Errorf("a failed ! removed a star the viewer already had: %+v", r)
	}
}

// bottomRow is the frame's last row: the help bar or a notice.
func bottomRow(m Model) string {
	lines := strings.Split(m.View(), "\n")
	return lines[len(lines)-1]
}

// topRows is the frame's first two rows, unstyled: the hop trail and the
// search line.
func topRows(m Model) (trail, search string) {
	lines := strings.Split(ansi.Strip(m.top()), "\n")
	return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[len(lines)-1])
}

// spinnerShows reports whether label shows in the content area, the
// spinner's place, rather than in the bottom row.
func spinnerShows(m Model, label string) bool {
	return strings.Contains(m.View(), label) && !strings.Contains(bottomRow(m), label)
}

func Test_SpinnerOnlyWhileTheNetworkIsBusy(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat()},
		readmes: map[string]string{"octocat/alpha": "# alpha readme"},
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	if !spinnerShows(m, "loading octocat…") {
		t.Errorf("no spinner in place while searching:\n%s", m.View())
	}
	m, _ = send(m, res)
	if strings.Contains(m.View(), "loading") {
		t.Errorf("spinner outlived the search:\n%s", m.View())
	}

	m = focusRepo(t, m, "A")
	m, msgs := send(m, keyRunes("r"))
	if !spinnerShows(m, "loading readme…") || !strings.Contains(m.View(), "alpha") {
		t.Errorf("no spinner over the dimmed profile while the readme loads:\n%s", m.View())
	}
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if strings.Contains(m.View(), "loading") {
		t.Errorf("spinner outlived the readme:\n%s", m.View())
	}

	// back and open it again: the cache answers, so no spinner at all
	m, _ = send(m, keyRunes("r"))
	m, msgs = send(m, keyRunes("r"))
	if strings.Contains(m.View(), "loading") {
		t.Errorf("spinner shown for a cached readme:\n%s", m.View())
	}
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if m.state.CurrentView != state.ReadmeView {
		t.Errorf("cached readme did not open")
	}
	if n := strings.Count(strings.Join(client.calls, "\n"), "readme octocat/alpha"); n != 1 {
		t.Errorf("fetched the readme %d times, want once", n)
	}
}

func Test_FailedReadmeIsQuietOnceTheUserMovesOn(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "A")

	m, msgs := send(m, keyRunes("r"))
	failed := find[commands.GetReadmeResponse](t, msgs)
	if failed.Err == nil {
		t.Fatal("expected the readme request to fail")
	}

	// the user moves on before the failure lands
	m, _ = send(m, keyRunes("j"))
	if strings.Contains(m.View(), "loading") {
		t.Errorf("spinner kept going after the user moved on:\n%s", m.View())
	}
	m, msgs = send(m, failed)
	if _, ok := firstOf[commands.SetMessage](msgs.all(t)); ok {
		t.Errorf("notice shown for a readme the user moved away from")
	}

	// without a keypress in between, the failure is reported
	m, msgs = send(m, keyRunes("r"))
	m, msgs = send(m, find[commands.GetReadmeResponse](t, msgs))
	m, _ = send(m, find[commands.SetMessage](t, msgs))
	if !strings.Contains(bottomRow(m), "couldn't load readme") {
		t.Fatalf("no notice for a failed readme: %q", bottomRow(m))
	}

	// and the next keypress clears it
	m, _ = send(m, keyRunes("k"))
	if strings.Contains(bottomRow(m), "couldn't load readme") {
		t.Errorf("notice survived a keypress: %q", bottomRow(m))
	}
}

func Test_StarIsInstantAndRevertsOnFailure(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat()},
		starErr: errors.New("You do not have permission to star this repository"),
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "A")
	stars := m.state.FocusedRepo().StarsCount

	m, msgs := send(m, keyRunes("s"))
	if r := m.state.FocusedRepo(); !r.ViewerHasStarred || r.StarsCount != stars+1 {
		t.Fatalf("star not shown at once: %+v", r)
	}

	m, msgs = send(m, find[commands.StarResponse](t, msgs))
	if r := m.state.FocusedRepo(); r.ViewerHasStarred || r.StarsCount != stars {
		t.Errorf("failed star not reverted: %+v", r)
	}
	m, _ = send(m, find[commands.SetMessage](t, msgs))
	if !strings.Contains(bottomRow(m), "star failed: You do not have permission") {
		t.Errorf("notice does not say why: %q", bottomRow(m))
	}
}

func Test_StarIgnoresPressesWhileInFlight(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "A")

	m, first := send(m, keyRunes("s"))
	starred := find[commands.StarResponse](t, first)
	m, second := send(m, keyRunes("s")) // before the star lands
	if _, ok := firstOf[commands.StarResponse](second.all(t)); ok || client.called("unstar A") {
		t.Fatalf("a second toggle went out while the first was in flight: calls %v", client.calls)
	}
	if r := m.state.FocusedRepo(); !r.ViewerHasStarred {
		t.Errorf("the ignored press flipped the card: %+v", r)
	}

	m, _ = send(m, starred)
	if r := m.state.FocusedRepo(); !r.ViewerHasStarred || r.StarsCount != 1 {
		t.Errorf("card does not match the server: %+v", r)
	}

	// once it lands, s toggles again
	_, msgs := send(m, keyRunes("s"))
	find[commands.StarResponse](t, msgs)
	if !client.called("unstar A") {
		t.Errorf("s after the response did not unstar: calls %v", client.calls)
	}
}

func Test_ReadmeHelpOffersStarOnlyForRepos(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": octocat()},
		readmes: map[string]string{"octocat/octocat": "# me", "octocat/alpha": "# alpha"},
	}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)

	m, _ = send(m, keyRunes("j")) // the profile
	m, msgs := send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if strings.Contains(bottomRow(m), "star repo") {
		t.Errorf("profile readme offers to star a repo: %q", bottomRow(m))
	}

	m, _ = send(m, keyRunes("r"))
	m = focusRepo(t, m, "A")
	m, msgs = send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if !strings.Contains(bottomRow(m), "star repo") {
		t.Errorf("repo readme does not offer to star it: %q", bottomRow(m))
	}
}
