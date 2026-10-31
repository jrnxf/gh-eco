package github

import (
	"context"
	"encoding/base64"
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
	// readmes caches decoded readmes by owner/name for the session.
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

// User fetches a profile, its pinned repositories and its contribution graph.
func (c *Client) User(ctx context.Context, login string) (models.User, error) {
	var query queries.GetUserQuery
	variables := map[string]interface{}{
		"login": queries.String(login),
		"first": queries.Int(6),
	}
	if err := c.gql.QueryWithContext(ctx, "GetUser", &query, variables); err != nil {
		return models.User{}, err
	}
	return mapUser(query), nil
}

// Readme fetches the decoded readme of owner/name. Readmes are cached for
// the session, so toggling back to one is instant.
func (c *Client) Readme(ctx context.Context, owner, name string) (string, error) {
	key := strings.ToLower(owner + "/" + name)
	c.mu.Lock()
	text, ok := c.readmes[key]
	c.mu.Unlock()
	if ok {
		return text, nil
	}

	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	path := fmt.Sprintf("repos/%s/%s/readme", url.PathEscape(owner), url.PathEscape(name))
	if err := c.rest.DoWithContext(ctx, "GET", path, nil, &response); err != nil {
		return "", err
	}
	text, err := decodeReadmeContent(response.Content, response.Encoding)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.readmes[key] = text
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
