package ui

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

// fakeClient serves canned users and records every call.
type fakeClient struct {
	mu      sync.Mutex
	calls   []string
	users   map[string]models.User
	readmes map[string]string
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

func (f *fakeClient) Readme(_ context.Context, owner, name string) (string, error) {
	f.record("readme " + owner + "/" + name)
	if r, ok := f.readmes[owner+"/"+name]; ok {
		return r, nil
	}
	return "", github.ErrNotFound
}

func (f *fakeClient) OrgReadme(_ context.Context, org string) (string, error) {
	f.record("org readme " + org)
	return "# " + org, nil
}

func (f *fakeClient) Star(_ context.Context, id string) (models.Starrable, error) {
	f.record("star " + id)
	return models.Starrable{Id: id, ViewerHasStarred: true, StargazerCount: 1}, nil
}

func (f *fakeClient) Unstar(_ context.Context, id string) (models.Starrable, error) {
	f.record("unstar " + id)
	return models.Starrable{Id: id}, nil
}

func octocat() models.User {
	return models.User{
		Login: "octocat",
		Url:   "https://github.com/octocat",
		PinnedRepos: []models.Repo{
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
	m, _ = send(m, keyRunes(login))
	m, msgs := send(m, enter)
	m, msgs = send(m, find[commands.SearchSubmitted](t, msgs))
	return m, find[commands.GetUserResponse](t, msgs)
}

func newModel(t *testing.T, client *fakeClient) Model {
	t.Helper()
	m, _ := send(New(client), tea.WindowSizeMsg{Width: 160, Height: 50})
	return m
}

func Test_StarResponseLandsAfterFocusMoved(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": octocat()}}
	m := newModel(t, client)
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)

	m, _ = send(m, keyRunes("j")) // user
	m, _ = send(m, keyRunes("j")) // alpha
	m, msgs := send(m, keyRunes("s"))
	starred := find[commands.StarResponse](t, msgs)
	if !client.called("star A") {
		t.Fatalf("s on alpha did not star it: calls %v", client.calls)
	}

	// move to beta before alpha's star lands
	m, _ = send(m, keyRunes("j"))
	m, _ = send(m, starred)

	if r := m.state.FocusedRepo(); r == nil || r.Id != "B" || r.ViewerHasStarred {
		t.Fatalf("focused repo = %+v, want unstarred beta", r)
	}
	if !m.state.User.PinnedRepos[0].ViewerHasStarred {
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
	for range 3 {
		m, _ = send(m, keyRunes("j")) // user, alpha, beta
	}

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
		m, _ = send(m, keyRunes("j")) // user
		m, _ = send(m, keyRunes("j")) // alpha
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
		PinnedRepos:    []models.Repo{{Id: "R_9", Name: "anvil", Owner: struct{ Login string }{"acme"}}},
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
	m, _ = send(m, keyRunes("j"))
	m, _ = send(m, keyRunes("j")) // alpha
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

func Test_ThanksOnlyAfterStarringGhEco(t *testing.T) {
	m := newModel(t, &fakeClient{})
	m.state.Mode = state.NormalMode // ! works once the search box is left

	m, msgs := send(m, keyRunes("!"))
	if _, ok := firstOf[commands.SetMessage](msgs.all(t)); ok {
		t.Errorf("thanked before the star request returned")
	}
	_, msgs = send(m, find[commands.StarResponse](t, msgs))
	if msg := find[commands.SetMessage](t, msgs); !strings.Contains(msg.Content, "tysm") {
		t.Errorf("notice after starring gh-eco = %q", msg.Content)
	}

	_, msgs = send(m, commands.StarResponse{Err: errors.New("boom")})
	if msg := find[commands.SetMessage](t, msgs); strings.Contains(msg.Content, "tysm") {
		t.Errorf("thanked for a star that failed")
	}
}

// bottomRow is the frame's last row: the help bar, a notice or a spinner.
func bottomRow(m Model) string {
	lines := strings.Split(m.View(), "\n")
	return lines[len(lines)-1]
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
	m, _ = send(m, keyRunes("j")) // alpha
	m, msgs = send(m, keyRunes("r"))
	m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
	if !strings.Contains(bottomRow(m), "star repo") {
		t.Errorf("repo readme does not offer to star it: %q", bottomRow(m))
	}
}
