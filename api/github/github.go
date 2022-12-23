package github

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	"github.com/jrnxf/gh-eco/api/github/mutations"
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/utils"
	graphql "github.com/shurcooL/graphql"
)

const GH_ECO_REPO_ID string = "R_kgDOHVAImQ"

func GetUser(login string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			return commands.GetUserResponse{Err: err}
		}

		var query queries.GetUserQuery

		variables := map[string]interface{}{
			"login": graphql.String(login),
			"first": graphql.Int(6),
		}
		err = client.Query("GetUser", &query, variables)
		if err != nil {
			return commands.GetUserResponse{Err: err}
		}
		return commands.GetUserResponse{User: utils.MapGetUserQueryToDisplayUser(query)}
	}
}

func GetReadme(name string, owner string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Println(err)
			return commands.GetReadmeResponse{Err: err}
		}

		var query queries.GetReadmeQuery

		variables := map[string]interface{}{
			"name":       graphql.String(name),
			"owner":      graphql.String(owner),
			"expression": graphql.String("HEAD:README.md"),
		}
		err = client.Query("GetReadme", &query, variables)
		if err != nil {
			log.Println(err)
			return commands.GetReadmeResponse{Err: err}
		}
		return commands.GetReadmeResponse{Readme: query.Repository.Object.Blob}
	}
}

func StarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Println(err)
			return commands.StarStarrableResponse{Err: err}
		}

		var mutation mutations.AddStarMutation

		variables := map[string]interface{}{
			"starrableId": graphql.ID(starrableId),
		}

		err = client.Mutate("StarStarrable", &mutation, variables)
		if err != nil {
			log.Println(err)
			return commands.StarStarrableResponse{Err: err}
		}
		return commands.StarStarrableResponse{Starrable: mutation.AddStar.Starrable}
	}
}

func RemoveStarStarrable(starrableId string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Println(err)
			return commands.RemoveStarStarrableResponse{Err: err}
		}

		var mutation mutations.RemoveStarMutation

		variables := map[string]interface{}{
			"starrableId": graphql.ID(starrableId),
		}

		err = client.Mutate("RemoveStarStarrable", &mutation, variables)
		if err != nil {
			log.Println(err)
			return commands.RemoveStarStarrableResponse{Err: err}
		}
		return commands.RemoveStarStarrableResponse{Starrable: mutation.RemoveStar.Starrable}
	}
}
