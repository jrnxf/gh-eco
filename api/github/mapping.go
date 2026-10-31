package github

import (
	"github.com/jrnxf/gh-eco/api/github/queries"
	"github.com/jrnxf/gh-eco/models"
)

// mapOwner maps a repositoryOwner lookup; ok is false when no user or
// organization has the login.
func mapOwner(query queries.GetOwnerQuery) (u models.User, ok bool) {
	owner := query.RepositoryOwner
	switch owner.Typename {
	case "User":
		return mapUser(owner.User), true
	case "Organization":
		return mapOrganization(owner.Organization), true
	}
	return models.User{}, false
}

func mapUser(qu queries.User) models.User {
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
		PinnedRepos:       mapPinned(qu.PinnedItems),
	}

	u.ActivityGraph.ContributionsCount = qu.ContributionsCollection.ContributionCalendar.TotalContributions
	for _, week := range qu.ContributionsCollection.ContributionCalendar.Weeks {
		u.ActivityGraph.Weeks = append(u.ActivityGraph.Weeks, week)
	}
	return u
}

func mapOrganization(qo queries.Organization) models.User {
	return models.User{
		IsOrganization:  true,
		Login:           qo.Login,
		Name:            qo.Name,
		Location:        qo.Location,
		Url:             qo.Url,
		Bio:             qo.Description,
		TwitterUsername: qo.TwitterUsername,
		WebsiteUrl:      qo.WebsiteUrl,
		PinnedRepos:     mapPinned(qo.PinnedItems),
	}
}

func mapPinned(items queries.PinnedItems) []models.Repo {
	var repos []models.Repo
	for _, node := range items.Nodes {
		r := node.Repository
		if r.Id == "" {
			// a pinned gist: only the Repository fragment is queried
			continue
		}
		repos = append(repos, models.Repo{
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
	return repos
}
