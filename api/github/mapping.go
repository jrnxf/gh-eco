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
	}
	u.Repos, u.TopRepos = mapRepos(qu.PinnedItems, qu.Repositories)

	u.ActivityGraph.ContributionsCount = qu.ContributionsCollection.ContributionCalendar.TotalContributions
	for _, week := range qu.ContributionsCollection.ContributionCalendar.Weeks {
		u.ActivityGraph.Weeks = append(u.ActivityGraph.Weeks, week)
	}
	return u
}

func mapOrganization(qo queries.Organization) models.User {
	u := models.User{
		IsOrganization:  true,
		Login:           qo.Login,
		Name:            qo.Name,
		Location:        qo.Location,
		Url:             qo.Url,
		Bio:             qo.Description,
		TwitterUsername: qo.TwitterUsername,
		WebsiteUrl:      qo.WebsiteUrl,
	}
	u.Repos, u.TopRepos = mapRepos(qo.PinnedItems, qo.Repositories)
	return u
}

// mapRepos returns the pinned repositories or, when nothing is pinned, the
// most-starred ones with top set.
func mapRepos(pinned queries.PinnedItems, starred queries.TopRepositories) (repos []models.Repo, top bool) {
	for _, node := range pinned.Nodes {
		if node.Repository.Id == "" {
			// a pinned gist: only the Repository fragment is queried
			continue
		}
		repos = append(repos, mapRepo(node.Repository))
	}
	if len(repos) > 0 {
		return repos, false
	}
	for _, r := range starred.Nodes {
		repos = append(repos, mapRepo(r))
	}
	return repos, len(repos) > 0
}

func mapRepo(r queries.Repository) models.Repo {
	return models.Repo{
		Id:               r.Id,
		Name:             r.Name,
		Description:      r.Description,
		StarsCount:       r.StargazerCount,
		ViewerHasStarred: r.ViewerHasStarred,
		Owner:            struct{ Login string }{Login: r.Owner.Login},
		Url:              r.Url,
		PrimaryLanguage:  r.PrimaryLanguage,
	}
}

func mapPeople(p queries.People) models.PeoplePage {
	page := models.PeoplePage{
		TotalCount:  p.TotalCount,
		HasNextPage: p.PageInfo.HasNextPage,
		EndCursor:   p.PageInfo.EndCursor,
	}
	for _, n := range p.Nodes {
		page.People = append(page.People, models.Person{Login: n.Login, Name: n.Name, Url: n.Url})
	}
	return page
}
