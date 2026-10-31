package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
)

// fakeGitHub serves canned responses and records every request it gets.
type fakeGitHub struct {
	t *testing.T
	// graphql answers GraphQL requests by operation name.
	graphql map[string]string
	// rest answers REST requests by path.
	rest map[string]string

	mu       sync.Mutex
	requests []recorded
}

type recorded struct {
	path      string
	query     string
	variables map[string]any
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := recorded{path: r.URL.Path}
	var body string
	if r.URL.Path == "/graphql" {
		var payload struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &payload); err != nil {
			f.t.Errorf("bad GraphQL payload: %v", err)
		}
		req.query, req.variables = payload.Query, payload.Variables
		for name, response := range f.graphql {
			if strings.HasPrefix(payload.Query, "query "+name+"(") ||
				strings.HasPrefix(payload.Query, "mutation "+name+"(") {
				body = response
			}
		}
	} else {
		body = f.rest[r.URL.Path]
	}
	f.mu.Lock()
	f.requests = append(f.requests, req)
	f.mu.Unlock()

	if body == "" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"Not Found"}`)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

// redirect sends every request to the fake server.
type redirect struct{ to *url.URL }

func (r redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = r.to.Scheme, r.to.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newFake(t *testing.T, fake *fakeGitHub) *Client {
	t.Helper()
	fake.t = t
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	to, _ := url.Parse(srv.URL)
	c, err := NewClientWithOptions(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
		Transport: redirect{to},
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

const userQuery = `query GetUser($first:Int!$login:String!){user(login: $login){id,login,name,location,url,bio,twitterUsername,isViewer,isFollowingViewer,viewerIsFollowing,websiteUrl,followers{totalCount},following{totalCount},pinnedItems(first: $first, types: [REPOSITORY]){nodes{... on Repository{id,name,description,stargazerCount,viewerHasStarred,url,owner{login},primaryLanguage{name,color}}}},contributionsCollection{contributionCalendar{totalContributions,weeks{contributionDays{contributionLevel}}}}}}`

// userResponse is shaped like a real response: a pinned gist matches no
// fragment and arrives as an empty object, and a repo can lack a language.
const userResponse = `{"data":{"user":{
  "id":"U_1","login":"octocat","name":"Mona","location":"Lisbon","url":"https://github.com/octocat",
  "bio":"hi","twitterUsername":"","isViewer":false,"isFollowingViewer":true,"viewerIsFollowing":false,
  "websiteUrl":"https://example.com","followers":{"totalCount":12},"following":{"totalCount":3},
  "pinnedItems":{"nodes":[
    {"id":"R_1","name":"hello-world","description":"first","stargazerCount":5,"viewerHasStarred":true,
     "url":"https://github.com/octocat/hello-world","owner":{"login":"octocat"},
     "primaryLanguage":{"name":"Go","color":"#00ADD8"}},
    {},
    {"id":"R_2","name":"notes","description":"","stargazerCount":0,"viewerHasStarred":false,
     "url":"https://github.com/octocat/notes","owner":{"login":"octocat"},"primaryLanguage":null}
  ]},
  "contributionsCollection":{"contributionCalendar":{"totalContributions":42,"weeks":[
    {"contributionDays":[{"contributionLevel":"NONE"},{"contributionLevel":"FOURTH_QUARTILE"}]}
  ]}}
}}}`

func Test_UserQueryContract(t *testing.T) {
	fake := &fakeGitHub{graphql: map[string]string{"GetUser": userResponse}}
	c := newFake(t, fake)

	u, err := c.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}

	req := fake.requests[0]
	if req.query != userQuery {
		t.Errorf("query changed:\n got %s\nwant %s", req.query, userQuery)
	}
	if req.variables["login"] != "octocat" || req.variables["first"] != float64(6) {
		t.Errorf("variables = %v", req.variables)
	}

	if u.Login != "octocat" || u.Name != "Mona" || u.FollowersCount != 12 || !u.IsFollowingViewer {
		t.Errorf("profile mapped wrong: %+v", u)
	}
	if len(u.PinnedRepos) != 2 || u.PinnedRepos[0].Id != "R_1" || u.PinnedRepos[1].Id != "R_2" {
		t.Fatalf("pinned repos = %+v, want hello-world and notes without the gist", u.PinnedRepos)
	}
	if r := u.PinnedRepos[0]; r.Owner.Login != "octocat" || r.PrimaryLanguage.Name != "Go" || !r.ViewerHasStarred || r.StarsCount != 5 {
		t.Errorf("repo mapped wrong: %+v", r)
	}
	if u.ActivityGraph.ContributionsCount != 42 || len(u.ActivityGraph.Weeks) != 1 ||
		u.ActivityGraph.Weeks[0].ContributionDays[1].ContributionLevel != "FOURTH_QUARTILE" {
		t.Errorf("activity graph mapped wrong: %+v", u.ActivityGraph)
	}
}

func Test_StarMutationContract(t *testing.T) {
	fake := &fakeGitHub{graphql: map[string]string{
		"AddStar":    `{"data":{"addStar":{"starrable":{"id":"R_1","stargazerCount":6,"viewerHasStarred":true}}}}`,
		"RemoveStar": `{"data":{"removeStar":{"starrable":{"id":"R_1","stargazerCount":5,"viewerHasStarred":false}}}}`,
	}}
	c := newFake(t, fake)

	s, err := c.Star(context.Background(), "R_1")
	if err != nil {
		t.Fatal(err)
	}
	if s.Id != "R_1" || !s.ViewerHasStarred || s.StargazerCount != 6 {
		t.Errorf("star result = %+v", s)
	}
	if s, err = c.Unstar(context.Background(), "R_1"); err != nil || s.ViewerHasStarred {
		t.Errorf("unstar result = %+v, %v", s, err)
	}

	want := []string{
		`mutation AddStar($input:AddStarInput!){addStar(input: $input){starrable{id,stargazerCount,viewerHasStarred}}}`,
		`mutation RemoveStar($input:RemoveStarInput!){removeStar(input: $input){starrable{id,stargazerCount,viewerHasStarred}}}`,
	}
	for i, req := range fake.requests {
		if req.query != want[i] {
			t.Errorf("mutation %d changed:\n got %s\nwant %s", i, req.query, want[i])
		}
		input, _ := req.variables["input"].(map[string]any)
		if input["starrableId"] != "R_1" {
			t.Errorf("mutation %d input = %v", i, req.variables["input"])
		}
	}
}

func Test_ReadmeIsFetchedOncePerRepo(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("# hello"))
	fake := &fakeGitHub{rest: map[string]string{
		"/repos/octocat/hello-world/readme": `{"content":"` + encoded + `","encoding":"base64"}`,
	}}
	c := newFake(t, fake)

	for range 3 {
		text, err := c.Readme(context.Background(), "octocat", "hello-world")
		if err != nil || text != "# hello" {
			t.Fatalf("readme = %q, %v", text, err)
		}
	}
	if _, err := c.Readme(context.Background(), "octocat", "missing"); err == nil {
		t.Errorf("missing readme returned no error")
	}
	if _, err := c.Readme(context.Background(), "octocat", "missing"); err == nil {
		t.Errorf("missing readme returned no error the second time")
	}

	if n := len(fake.requests); n != 3 {
		t.Errorf("made %d requests, want 1 for the cached readme and 2 for the uncached miss", n)
	}
}
