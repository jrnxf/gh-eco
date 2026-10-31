package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jrnxf/gh-eco/api/github/mutations"
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/models"
	ghv4 "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GhEcoRepoID is the node id of the jrnxf/gh-eco repository.
const GhEcoRepoID = "R_kgDOHVAImQ"

// Client talks to the GitHub API with the credentials of the gh CLI.
type Client struct {
	gql  *ghv4.Client
	rest *api.RESTClient
}

// NewClient builds a Client from the gh CLI's authentication.
func NewClient() (*Client, error) {
	output, _, err := gh.Exec("auth", "token")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve gh auth token: %w", err)
	}
	token := strings.TrimSpace(output.String())
	if token == "" {
		return nil, fmt.Errorf("empty gh auth token")
	}
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return &Client{
		gql:  ghv4.NewClient(oauth2.NewClient(context.Background(), src)),
		rest: rest,
	}, nil
}

// User fetches a profile, its pinned repositories and its contribution graph.
func (c *Client) User(ctx context.Context, login string) (models.User, error) {
	var query queries.GetUserQuery
	variables := map[string]interface{}{
		"login": ghv4.String(login),
		"first": ghv4.Int(6),
	}
	if err := c.gql.Query(ctx, &query, variables); err != nil {
		return models.User{}, err
	}
	return mapUser(query), nil
}

// Readme fetches the decoded readme of owner/name.
func (c *Client) Readme(ctx context.Context, owner, name string) (string, error) {
	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	path := fmt.Sprintf("repos/%s/%s/readme", url.PathEscape(owner), url.PathEscape(name))
	if err := c.rest.DoWithContext(ctx, "GET", path, nil, &response); err != nil {
		return "", err
	}
	return decodeReadmeContent(response.Content, response.Encoding)
}

// Star stars the starrable with the given node id.
func (c *Client) Star(ctx context.Context, id string) (models.Starrable, error) {
	var mutation mutations.AddStarMutation
	input := ghv4.AddStarInput{StarrableID: ghv4.ID(id)}
	if err := c.gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return models.Starrable{}, err
	}
	return models.Starrable(mutation.AddStar.Starrable), nil
}

// Unstar removes the viewer's star from the starrable with the given node id.
func (c *Client) Unstar(ctx context.Context, id string) (models.Starrable, error) {
	var mutation mutations.RemoveStarMutation
	input := ghv4.RemoveStarInput{StarrableID: ghv4.ID(id)}
	if err := c.gql.Mutate(ctx, &mutation, input, nil); err != nil {
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
