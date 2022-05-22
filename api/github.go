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
			Repo Repo `graphql:"... on Repository"`
		}
	} `graphql:"pinnedItems(first: $first)"`
	ContributionsCollection struct {
		ContributionCalendar struct {
			TotalContributions int
			Weeks              []WeeklyContribution
		}
	}
}

type Repo struct {
	Id              string
	Name            string
	Description     string
	StargazerCount  int
	Url             string
	PrimaryLanguage struct {
		Name  string
		Color string
	}
}

type WeeklyContribution struct {
	ContributionDays []struct {
		ContributionLevel string
	}
}

type TotalCount struct {
	TotalCount int
}

type SearchUserResponse struct {
	Err  error
	User User
}

// mutation MyMutation {
// 	removeStar(input: {starrableId: "R_kgDOGuzvkw"}) {
// 	  starrable {
// 		viewerHasStarred
// 		stargazerCount
// 	  }
// 	}
//   }

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
			"first": graphql.Int(6),
		}
		log.Println("searching for", login)
		err = client.Query("SearchUser", &query, variables)
		if err != nil {
			log.Println(err)
			return SearchUserResponse{Err: err}
		}
		log.Println("found ", query.User.Name)
		return SearchUserResponse{User: query.User}
	}

}