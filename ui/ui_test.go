package ui

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

var errNotFound = errors.New("not found")

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
	return models.User{}, errNotFound
}

func (f *fakeClient) Readme(_ context.Context, owner, name string) (string, error) {
	f.record("readme " + owner + "/" + name)
	if r, ok := f.readmes[owner+"/"+name]; ok {
		return r, nil
	}
	return "", errNotFound
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

// drain runs cmd and every command it batches, and returns the messages
// that arrive within a short window. The fake client answers at once;
// cursor, spinner and timer ticks sleep far longer and are left behind.
func drain(cmd tea.Cmd) []tea.Msg {
	out := make(chan tea.Msg, 64)
	var run func(tea.Cmd)
	run = func(c tea.Cmd) {
		if c == nil {
			return
		}
		go func() {
			msg := c()
			if batch, ok := msg.(tea.BatchMsg); ok {
				for _, c := range batch {
					run(c)
				}
				return
			}
			out <- msg
		}()
	}
	run(cmd)

	var msgs []tea.Msg
	deadline := time.After(50 * time.Millisecond)
	for {
		select {
		case msg := <-out:
			msgs = append(msgs, msg)
		case <-deadline:
			return msgs
		}
	}
}

// find returns the first message of type T.
func find[T tea.Msg](t *testing.T, msgs []tea.Msg) T {
	t.Helper()
	for _, msg := range msgs {
		if m, ok := msg.(T); ok {
			return m
		}
	}
	var zero T
	t.Fatalf("no %T among %d messages", zero, len(msgs))
	return zero
}

// send feeds msg to m and returns the updated model and the messages its
// commands produced.
func send(m Model, msg tea.Msg) (Model, []tea.Msg) {
	next, cmd := m.Update(msg)
	return next.(Model), drain(cmd)
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

	_, _ = send(m, keyRunes("s"))
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
