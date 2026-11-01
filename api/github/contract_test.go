package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/jrnxf/gh-eco/models"
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

const ownerQuery = `query GetOwner($first:Int!$login:String!){repositoryOwner(login: $login){__typename,... on User{login,name,location,url,bio,twitterUsername,isViewer,isFollowingViewer,viewerIsFollowing,websiteUrl,followers{totalCount},following{totalCount},pinnedItems(first: $first, types: [REPOSITORY]){nodes{... on Repository{id,name,description,stargazerCount,viewerHasStarred,url,owner{login},primaryLanguage{name,color}}}},repositories(first: $first, privacy: PUBLIC, isFork: false, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC}){nodes{id,name,description,stargazerCount,viewerHasStarred,url,owner{login},primaryLanguage{name,color}}},contributionsCollection{contributionCalendar{totalContributions,weeks{contributionDays{contributionLevel}}}}},... on Organization{login,name,location,url,description,twitterUsername,websiteUrl,pinnedItems(first: $first, types: [REPOSITORY]){nodes{... on Repository{id,name,description,stargazerCount,viewerHasStarred,url,owner{login},primaryLanguage{name,color}}}},repositories(first: $first, privacy: PUBLIC, isFork: false, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC}){nodes{id,name,description,stargazerCount,viewerHasStarred,url,owner{login},primaryLanguage{name,color}}}}}}`

// userResponse is shaped like a real response: a pinned gist matches no
// fragment and arrives as an empty object, and a repo can lack a language.
const userResponse = `{"data":{"repositoryOwner":{
  "__typename":"User","login":"octocat","name":"Mona","location":"Lisbon","url":"https://github.com/octocat",
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
  "repositories":{"nodes":[
    {"id":"R_3","name":"popular","description":"","stargazerCount":900,"viewerHasStarred":false,
     "url":"https://github.com/octocat/popular","owner":{"login":"octocat"},"primaryLanguage":null}
  ]},
  "contributionsCollection":{"contributionCalendar":{"totalContributions":42,"weeks":[
    {"contributionDays":[{"contributionLevel":"NONE"},{"contributionLevel":"FOURTH_QUARTILE"}]}
  ]}}
}}}`

func Test_UserLookupContract(t *testing.T) {
	fake := &fakeGitHub{graphql: map[string]string{"GetOwner": userResponse}}
	c := newFake(t, fake)

	u, err := c.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}

	req := fake.requests[0]
	if req.query != ownerQuery {
		t.Errorf("query changed:\n got %s\nwant %s", req.query, ownerQuery)
	}
	if req.variables["login"] != "octocat" || req.variables["first"] != float64(6) {
		t.Errorf("variables = %v", req.variables)
	}

	if u.Login != "octocat" || u.Name != "Mona" || u.FollowersCount != 12 || !u.IsFollowingViewer {
		t.Errorf("profile mapped wrong: %+v", u)
	}
	if len(u.Repos) != 2 || u.Repos[0].Id != "R_1" || u.Repos[1].Id != "R_2" || u.TopRepos {
		t.Fatalf("repos = %+v, top = %v; want the pins hello-world and notes without the gist", u.Repos, u.TopRepos)
	}
	if r := u.Repos[0]; r.Owner.Login != "octocat" || r.PrimaryLanguage.Name != "Go" || !r.ViewerHasStarred || r.StarsCount != 5 {
		t.Errorf("repo mapped wrong: %+v", r)
	}
	if u.ActivityGraph.ContributionsCount != 42 || len(u.ActivityGraph.Weeks) != 1 ||
		u.ActivityGraph.Weeks[0].ContributionDays[1].ContributionLevel != "FOURTH_QUARTILE" {
		t.Errorf("activity graph mapped wrong: %+v", u.ActivityGraph)
	}
}

func Test_OrganizationLookup(t *testing.T) {
	fake := &fakeGitHub{graphql: map[string]string{"GetOwner": `{"data":{"repositoryOwner":{
	  "__typename":"Organization","login":"acme","name":"Acme","description":"we make things",
	  "location":"","url":"https://github.com/acme","twitterUsername":"acme","websiteUrl":"",
	  "pinnedItems":{"nodes":[{"id":"R_9","name":"anvil","description":"","stargazerCount":7,
	    "viewerHasStarred":false,"url":"https://github.com/acme/anvil","owner":{"login":"acme"},
	    "primaryLanguage":{"name":"Rust","color":"#dea584"}}]}}}}`}}
	c := newFake(t, fake)

	u, err := c.User(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsOrganization || u.Login != "acme" || u.Bio != "we make things" || u.TwitterUsername != "acme" {
		t.Errorf("organization mapped wrong: %+v", u)
	}
	if len(u.Repos) != 1 || u.Repos[0].Name != "anvil" {
		t.Errorf("pinned repos = %+v", u.Repos)
	}
}

func Test_TopReposWhenNothingIsPinned(t *testing.T) {
	// only a pinned gist, which the Repository fragment leaves empty
	fake := &fakeGitHub{graphql: map[string]string{"GetOwner": `{"data":{"repositoryOwner":{
	  "__typename":"User","login":"octocat","followers":{"totalCount":0},"following":{"totalCount":0},
	  "pinnedItems":{"nodes":[{}]},
	  "repositories":{"nodes":[
	    {"id":"R_7","name":"famous","description":"","stargazerCount":900,"viewerHasStarred":false,
	     "url":"https://github.com/octocat/famous","owner":{"login":"octocat"},"primaryLanguage":null},
	    {"id":"R_8","name":"known","description":"","stargazerCount":40,"viewerHasStarred":true,
	     "url":"https://github.com/octocat/known","owner":{"login":"octocat"},"primaryLanguage":null}
	  ]},
	  "contributionsCollection":{"contributionCalendar":{"totalContributions":0,"weeks":[]}}
	}}}`}}
	c := newFake(t, fake)

	u, err := c.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if !u.TopRepos || len(u.Repos) != 2 || u.Repos[0].Name != "famous" || u.Repos[1].StarsCount != 40 || !u.Repos[1].ViewerHasStarred {
		t.Errorf("repos = %+v, top = %v; want famous then known, marked as top repos", u.Repos, u.TopRepos)
	}
}

func Test_NoReposAtAll(t *testing.T) {
	fake := &fakeGitHub{graphql: map[string]string{"GetOwner": `{"data":{"repositoryOwner":{
	  "__typename":"Organization","login":"acme","pinnedItems":{"nodes":[]},"repositories":{"nodes":[]}}}}`}}
	c := newFake(t, fake)

	u, err := c.User(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(u.Repos) != 0 || u.TopRepos {
		t.Errorf("repos = %+v, top = %v; want none, not labelled as top repos", u.Repos, u.TopRepos)
	}
}

func Test_PeopleContract(t *testing.T) {
	page := func(logins ...string) string {
		var nodes []string
		for _, l := range logins {
			nodes = append(nodes, `{"login":"`+l+`","name":"","url":"https://github.com/`+l+`"}`)
		}
		return `"totalCount":3,"pageInfo":{"hasNextPage":true,"endCursor":"CUR"},"nodes":[` + strings.Join(nodes, ",") + `]`
	}
	fake := &fakeGitHub{graphql: map[string]string{
		"GetFollowers": `{"data":{"user":{"followers":{` + page("alice", "bob") + `}}}}`,
		"GetFollowing": `{"data":{"user":{"following":{` + page("carol") + `}}}}`,
	}}
	c := newFake(t, fake)

	followers, err := c.People(context.Background(), "octocat", models.Followers, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(followers.People) != 2 || followers.People[1].Login != "bob" || followers.People[1].Url != "https://github.com/bob" || followers.TotalCount != 3 ||
		!followers.HasNextPage || followers.EndCursor != "CUR" {
		t.Errorf("followers page mapped wrong: %+v", followers)
	}
	following, err := c.People(context.Background(), "octocat", models.Following, "CUR")
	if err != nil || len(following.People) != 1 || following.People[0].Login != "carol" {
		t.Errorf("following page = %+v, %v", following, err)
	}

	want := []string{
		`query GetFollowers($after:String$first:Int!$login:String!){user(login: $login){followers(first: $first, after: $after){totalCount,pageInfo{hasNextPage,endCursor},nodes{login,name,url}}}}`,
		`query GetFollowing($after:String$first:Int!$login:String!){user(login: $login){following(first: $first, after: $after){totalCount,pageInfo{hasNextPage,endCursor},nodes{login,name,url}}}}`,
	}
	for i, req := range fake.requests {
		if req.query != want[i] {
			t.Errorf("query %d changed:\n got %s\nwant %s", i, req.query, want[i])
		}
		if req.variables["login"] != "octocat" || req.variables["first"] != float64(50) {
			t.Errorf("query %d variables = %v", i, req.variables)
		}
	}
	// the first page sends a null cursor, later pages the previous end
	if after, ok := fake.requests[0].variables["after"]; !ok || after != nil {
		t.Errorf("first page after = %v (sent %v), want null", after, ok)
	}
	if fake.requests[1].variables["after"] != "CUR" {
		t.Errorf("next page after = %v, want CUR", fake.requests[1].variables["after"])
	}
}

func Test_UnknownLoginIsNotFound(t *testing.T) {
	// GitHub answers an unknown login with a null owner, not an error
	fake := &fakeGitHub{graphql: map[string]string{"GetOwner": `{"data":{"repositoryOwner":null}}`}}
	c := newFake(t, fake)

	if _, err := c.User(context.Background(), "ghost"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func Test_ServerErrorIsNotNotFound(t *testing.T) {
	c := newFake(t, &fakeGitHub{}) // answers 404 to everything
	_, err := c.User(context.Background(), "octocat")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want a transport error distinct from ErrNotFound", err)
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
		"/repos/octocat/hello-world/readme":              `{"content":"` + encoded + `","encoding":"base64"}`,
		"/repos/acme/.github/contents/profile/README.md": `{"content":"` + encoded + `","encoding":"base64"}`,
	}}
	c := newFake(t, fake)

	if _, ok := c.CachedReadme("octocat", "hello-world"); ok {
		t.Errorf("readme cached before it was fetched")
	}
	for range 3 {
		text, err := c.Readme(context.Background(), "octocat", "hello-world")
		if err != nil || text != "# hello" {
			t.Fatalf("readme = %q, %v", text, err)
		}
	}
	if text, ok := c.CachedReadme("OctoCat", "Hello-World"); !ok || text != "# hello" {
		t.Errorf("cached readme = %q, %v; lookups ignore case like GitHub", text, ok)
	}
	if _, err := c.Readme(context.Background(), "octocat", "missing"); err == nil {
		t.Errorf("missing readme returned no error")
	}
	if _, err := c.Readme(context.Background(), "octocat", "missing"); err == nil {
		t.Errorf("missing readme returned no error the second time")
	}

	if _, err := c.OrgReadme(context.Background(), "acme"); err != nil {
		t.Errorf("org readme: %v", err)
	}
	if _, ok := c.CachedOrgReadme("acme"); !ok {
		t.Errorf("org readme not cached")
	}
	if _, ok := c.CachedReadme("octocat", "missing"); ok {
		t.Errorf("a failed fetch was cached")
	}
	if fake.requests[len(fake.requests)-1].path != "/repos/acme/.github/contents/profile/README.md" {
		t.Errorf("org readme fetched from %s", fake.requests[len(fake.requests)-1].path)
	}

	if n := len(fake.requests); n != 4 {
		t.Errorf("made %d requests, want 1 for the cached readme, 2 for the uncached miss and 1 for the org", n)
	}
}
