package api

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	graphql "github.com/shurcooL/graphql"
)

type User struct {
	Login    string
	Name     string
	Location string
	// Email             string // requires bigger token
	Bio               string
	Company           string
	TwitterUsername   string
	WebsiteUrl        string
	ViewerIsFollowing bool
	IsFollowingViewer bool
	IsViewer          bool
	IsHireable        bool
	Status            struct {
		Emoji   string
		Message string
	}
	RepositoriesContributedTo TotalCount
	Followers                 TotalCount
	Following                 TotalCount
	PinnedItems               struct {
		Nodes []struct {
			Repo struct {
				Name           string
				StargazerCount int
				Url            string
			} `graphql:"... on Repository"`
		}
	} `graphql:"pinnedItems(first: $first)"`
}

type TotalCount struct {
	TotalCount int
}

type SearchUserResponse struct {
	Err  error
	User User
}

func SearchUser(login string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			return SearchUserResponse{Err: err}
		}

		var query struct {
			User User `graphql:"user(login: $login)"`
		}

		variables := map[string]interface{}{
			"login": graphql.String(login),
			"first": graphql.Int(3),
		}
		err = client.Query("SearchUser", &query, variables)
		if err != nil {
			log.Fatal(err)
			return SearchUserResponse{Err: err}
		}

		return SearchUserResponse{User: query.User}
	}

}
