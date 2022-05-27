package api

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	"github.com/coloradocolby/gh-eco/types/display"
	"github.com/coloradocolby/gh-eco/types/queries"
	"github.com/coloradocolby/gh-eco/utils"
	graphql "github.com/shurcooL/graphql"
)

type GetUserResponse struct {
	Err  error
	User display.User
}

func GetUser(login string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			return GetUserResponse{Err: err}
		}

		var query queries.GetUser

		variables := map[string]interface{}{
			"login": graphql.String(login),
			"first": graphql.Int(6),
		}
		log.Println("GetUser START")
		err = client.Query("GetUser", &query, variables)
		if err != nil {
			log.Println(err)
			return GetUserResponse{Err: err}
		}
		log.Println("GetUser END")
		return GetUserResponse{User: utils.MapGetUserQueryToDisplayUser(query)}
	}
}

type GetReadmeResponse struct {
	Err    error
	Readme display.Blob
}

func GetReadme(name string, owner string) tea.Cmd {
	return func() tea.Msg {
		log.Println("GR")
		client, err := gh.GQLClient(nil)
		if err != nil {
			return GetReadmeResponse{Err: err}
		}

		var query queries.GetReadme

		variables := map[string]interface{}{
			"name":       graphql.String(name),
			"owner":      graphql.String(owner),
			"expression": graphql.String("HEAD:README.md"),
		}
		log.Println("GetReadme START")

		err = client.Query("GetReadme", &query, variables)
		if err != nil {
			log.Println(err)
			return GetReadmeResponse{Err: err}
		}
		log.Println("GetReadme END")
		return GetReadmeResponse{Readme: query.Repository.Object.Blob}
	}
}
