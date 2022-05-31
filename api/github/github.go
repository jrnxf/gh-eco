package github

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh"
	"github.com/coloradocolby/gh-eco/api/github/mutations"
	"github.com/coloradocolby/gh-eco/api/github/queries"
	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/models"
	"github.com/coloradocolby/gh-eco/utils"
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

func FollowUser(userId string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Println(err)
			return commands.FollowUserResponse{Err: err}
		}

		var mutation mutations.FollowUserMutation

		variables := map[string]interface{}{
			"userId": graphql.ID(userId),
		}

		err = client.Mutate("FollowUser", &mutation, variables)
		if err != nil {
			log.Println(err)
			return commands.FollowUserResponse{Err: err}
		}

		res := mutation.FollowUser.User
		return commands.FollowUserResponse{User: models.User{
			Id:                res.Id,
			ViewerIsFollowing: res.ViewerIsFollowing,
			FollowersCount:    res.Followers.TotalCount,
		}}
	}
}

func UnfollowUser(userId string) tea.Cmd {
	return func() tea.Msg {
		client, err := gh.GQLClient(nil)
		if err != nil {
			log.Println(err)
			return commands.UnfollowUserResponse{Err: err}
		}

		var mutation mutations.UnfollowUserMutation

		variables := map[string]interface{}{
			"userId": graphql.ID(userId),
		}

		err = client.Mutate("UnfollowUser", &mutation, variables)
		if err != nil {
			log.Println(err)
			return commands.UnfollowUserResponse{Err: err}
		}

		res := mutation.UnfollowUser.User
		return commands.UnfollowUserResponse{User: models.User{
			Id:                res.Id,
			ViewerIsFollowing: res.ViewerIsFollowing,
			FollowersCount:    res.Followers.TotalCount,
		}}
	}
}
