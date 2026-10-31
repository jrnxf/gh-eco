package github

import (
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/models"
)

func mapUser(query queries.GetUserQuery) models.User {
	qu := query.User
	u := models.User{
		Login:             qu.Login,
		Name:              qu.Name,
		Location:          qu.Location,
		Url:               qu.Url,
		Bio:               qu.Bio,
		TwitterUsername:   qu.TwitterUsername,
		WebsiteUrl:        qu.WebsiteUrl,
		FollowersCount:    qu.Followers.TotalCount,
		FollowingCount:    qu.Following.TotalCount,
		IsViewer:          qu.IsViewer,
		IsFollowingViewer: qu.IsFollowingViewer,
		ViewerIsFollowing: qu.ViewerIsFollowing,
	}

	u.ActivityGraph.ContributionsCount = qu.ContributionsCollection.ContributionCalendar.TotalContributions
	for _, week := range qu.ContributionsCollection.ContributionCalendar.Weeks {
		u.ActivityGraph.Weeks = append(u.ActivityGraph.Weeks, week)
	}

	for _, node := range qu.PinnedItems.Nodes {
		r := node.Repository
		if r.Id == "" {
			// a pinned gist: only the Repository fragment is queried
			continue
		}
		u.PinnedRepos = append(u.PinnedRepos, models.Repo{
			Id:               r.Id,
			Name:             r.Name,
			Description:      r.Description,
			StarsCount:       r.StargazerCount,
			ViewerHasStarred: r.ViewerHasStarred,
			Owner:            struct{ Login string }{Login: r.Owner.Login},
			Url:              r.Url,
			PrimaryLanguage:  r.PrimaryLanguage,
		})
	}

	return u
}
