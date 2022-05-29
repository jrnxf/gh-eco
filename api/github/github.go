package github

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	"github.com/coloradocolby/gh-eco/api/github/mutations"
	"github.com/coloradocolby/gh-eco/api/github/queries"
	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/utils"
	graphql "github.com/shurcooL/graphql"
)

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
		log.Println("GetReadme START")
		err = client.Query("GetReadme", &query, variables)
		log.Println("GetReadme END")
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

		var mutation mutations.StarMutation

		variables := map[string]interface{}{
			"starrableId": graphql.ID(starrableId),
		}

		log.Println("StarStarrable START")
		err = client.Mutate("StarStarrable", &mutation, variables)
		log.Println("StarStarrable END")
		if err != nil {
			log.Println(err)
			return commands.StarStarrableResponse{Err: err}
		}
		log.Println(mutation.AddStar.Starrable)
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

		log.Println("RemoveStarStarrable START")
		err = client.Mutate("RemoveStarStarrable", &mutation, variables)
		log.Println("RemoveStarStarrable END")
		if err != nil {
			log.Println(err)
			return commands.RemoveStarStarrableResponse{Err: err}
		}
		log.Println(mutation.RemoveStar.Starrable)
		return commands.RemoveStarStarrableResponse{Starrable: mutation.RemoveStar.Starrable}
	}
}
