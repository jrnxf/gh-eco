package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jrnxf/gh-eco/api/github/mutations"
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/models"
)

// GhEcoRepoID is the node id of the jrnxf/gh-eco repository.
const GhEcoRepoID = "R_kgDOHVAImQ"

// requestTimeout bounds every API request, so a stalled connection ends in
// an error instead of a spinner that never stops.
const requestTimeout = 15 * time.Second

// Client talks to the GitHub API with the credentials of the gh CLI, on the
// host gh is configured for.
type Client struct {
	gql  *api.GraphQLClient
	rest *api.RESTClient

	mu sync.Mutex
	// readmes caches decoded readmes by request path for the session.
	readmes map[string]string
}

// NewClient builds a Client from the gh CLI's configuration.
func NewClient() (*Client, error) {
	return NewClientWithOptions(api.ClientOptions{})
}

// NewClientWithOptions builds a Client with explicit go-gh options; unset
// options are resolved from the gh CLI's configuration.
func NewClientWithOptions(opts api.ClientOptions) (*Client, error) {
	if opts.Timeout == 0 {
		opts.Timeout = requestTimeout
	}
	gql, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, err
	}
	rest, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, err
	}
	return &Client{gql: gql, rest: rest, readmes: map[string]string{}}, nil
}

// ErrNotFound means no user or organization has the login.
var ErrNotFound = errors.New("not found")

// peoplePageSize is how many followers or followed people one request
// fetches.
const peoplePageSize = 50

// User fetches a user or organization profile with its pinned repositories,
// or its most-starred ones when nothing is pinned, and a user's
// contribution graph.
func (c *Client) User(ctx context.Context, login string) (models.User, error) {
	var query queries.GetOwnerQuery
	variables := map[string]interface{}{
		"login": queries.String(login),
		"first": queries.Int(6),
	}
	if err := c.gql.QueryWithContext(ctx, "GetOwner", &query, variables); err != nil {
		return models.User{}, err
	}
	u, ok := mapOwner(query)
	if !ok {
		return models.User{}, ErrNotFound
	}
	return u, nil
}

// People fetches one page of a user's followers or of the people they
// follow, continuing after the cursor after, or from the start when it is
// empty.
func (c *Client) People(ctx context.Context, login string, kind models.PeopleKind, after string) (models.PeoplePage, error) {
	var cursor *queries.String
	if after != "" {
		cursor = (*queries.String)(&after)
	}
	variables := map[string]interface{}{
		"login": queries.String(login),
		"first": queries.Int(peoplePageSize),
		"after": cursor,
	}
	if kind == models.Following {
		var query queries.GetFollowingQuery
		if err := c.gql.QueryWithContext(ctx, "GetFollowing", &query, variables); err != nil {
			return models.PeoplePage{}, err
		}
		return mapPeople(query.User.Following), nil
	}
	var query queries.GetFollowersQuery
	if err := c.gql.QueryWithContext(ctx, "GetFollowers", &query, variables); err != nil {
		return models.PeoplePage{}, err
	}
	return mapPeople(query.User.Followers), nil
}

// Readme fetches the decoded readme of owner/name.
func (c *Client) Readme(ctx context.Context, owner, name string) (string, error) {
	return c.cachedContent(ctx, readmePath(owner, name))
}

// OrgReadme fetches an organization's profile readme, which lives in its
// .github repository.
func (c *Client) OrgReadme(ctx context.Context, org string) (string, error) {
	return c.cachedContent(ctx, orgReadmePath(org))
}

// CachedReadme returns the readme of owner/name if it is already cached.
func (c *Client) CachedReadme(owner, name string) (string, bool) {
	return c.cached(readmePath(owner, name))
}

// CachedOrgReadme returns an organization's profile readme if it is
// already cached.
func (c *Client) CachedOrgReadme(org string) (string, bool) {
	return c.cached(orgReadmePath(org))
}

func readmePath(owner, name string) string {
	return fmt.Sprintf("repos/%s/%s/readme", url.PathEscape(owner), url.PathEscape(name))
}

func orgReadmePath(org string) string {
	return fmt.Sprintf("repos/%s/.github/contents/profile/README.md", url.PathEscape(org))
}

func (c *Client) cached(path string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	text, ok := c.readmes[strings.ToLower(path)]
	return text, ok
}

// cachedContent fetches and decodes a REST file response. Results are
// cached for the session, so toggling back to a readme is instant.
func (c *Client) cachedContent(ctx context.Context, path string) (string, error) {
	if text, ok := c.cached(path); ok {
		return text, nil
	}

	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.rest.DoWithContext(ctx, "GET", path, nil, &response); err != nil {
		return "", err
	}
	text, err := decodeReadmeContent(response.Content, response.Encoding)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.readmes[strings.ToLower(path)] = text
	c.mu.Unlock()
	return text, nil
}

// Star stars the starrable with the given node id.
func (c *Client) Star(ctx context.Context, id string) (models.Starrable, error) {
	var mutation mutations.AddStarMutation
	variables := map[string]interface{}{
		"input": mutations.AddStarInput{StarrableID: id},
	}
	if err := c.gql.MutateWithContext(ctx, "AddStar", &mutation, variables); err != nil {
		return models.Starrable{}, err
	}
	return models.Starrable(mutation.AddStar.Starrable), nil
}

// Unstar removes the viewer's star from the starrable with the given node id.
func (c *Client) Unstar(ctx context.Context, id string) (models.Starrable, error) {
	var mutation mutations.RemoveStarMutation
	variables := map[string]interface{}{
		"input": mutations.RemoveStarInput{StarrableID: id},
	}
	if err := c.gql.MutateWithContext(ctx, "RemoveStar", &mutation, variables); err != nil {
		return models.Starrable{}, err
	}
	return models.Starrable(mutation.RemoveStar.Starrable), nil
}

// decodeReadmeContent decodes the content field of the REST readme
// response. GitHub returns base64 with embedded newlines, which the
// strict std decoder rejects, so strip them first.
func decodeReadmeContent(content, encoding string) (string, error) {
	if encoding != "base64" {
		return "", fmt.Errorf("unexpected readme encoding %q", encoding)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
