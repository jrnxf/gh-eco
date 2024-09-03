package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/cli/go-gh/v2"
	"github.com/jrnxf/gh-eco/api/github/mutations"
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/utils"
	ghv4 "github.com/shurcooL/githubv4"
	graphql "github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

const GH_ECO_REPO_ID string = "R_kgDOHVAImQ"

var (
	clientInstance *ghv4.Client
	once           sync.Once
)

// GetClient initializes a GitHub GraphQL client instance with a token obtained from GitHub CLI.
func GetClient() *ghv4.Client {
	once.Do(func() {
		output, _, err := gh.Exec("auth", "token")
		if err != nil {
			fmt.Println("Unable to retrieve access token")
		}

		token := strings.TrimSpace(output.String())
		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient := oauth2.NewClient(context.Background(), src)
		clientInstance = ghv4.NewClient(httpClient)
	})

	return clientInstance
}

func GetUser(login string) tea.Cmd {
	return func() tea.Msg {
		client := GetClient()

		var query queries.GetUserQuery

		variables := map[string]interface{}{
			"login": graphql.String(login),
			"first": graphql.Int(6),
		}

		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			fmt.Println(err.Error())
			return commands.GetUserResponse{Err: err}
		}

		return commands.GetUserResponse{User: utils.MapGetUserQueryToDisplayUser(query)}
	}
}

func GetReadme(name string, owner string) tea.Cmd {
	return func() tea.Msg {
		client := GetClient()

		var query queries.GetReadmeQuery

		variables := map[string]interface{}{
			"name":       graphql.String(name),
			"owner":      graphql.String(owner),
			"expression": graphql.String("HEAD:README.md"),
		}

		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			log.Println(err)
			return commands.GetReadmeResponse{Err: err}
		}

		return commands.GetReadmeResponse{Readme: query.Repository.Object.Blob}
	}
}

func StarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client := GetClient()

		var mutation mutations.AddStarMutation

		variables := map[string]interface{}{
			"starrableId": graphql.ID(starrableId),
		}

		err := client.Mutate(context.Background(), &mutation, variables, nil)
		if err != nil {
			log.Println(err)
			return commands.StarStarrableResponse{Err: err}
		}

		return commands.StarStarrableResponse{Starrable: mutation.AddStar.Starrable}
	}
}

func RemoveStarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client := GetClient()

		var mutation mutations.RemoveStarMutation

		variables := map[string]interface{}{
			"starrableId": graphql.ID(starrableId),
		}

		err := client.Mutate(context.Background(), &mutation, variables, nil)
		if err != nil {
			log.Println(err)
			return commands.RemoveStarStarrableResponse{Err: err}
		}

		return commands.RemoveStarStarrableResponse{Starrable: mutation.RemoveStar.Starrable}
	}
}
