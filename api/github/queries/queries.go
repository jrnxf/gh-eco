package queries

// GetOwnerQuery looks up a user or an organization by login.
type GetOwnerQuery struct {
	RepositoryOwner struct {
		Typename     string       `graphql:"__typename"`
		User         User         `graphql:"... on User"`
		Organization Organization `graphql:"... on Organization"`
	} `graphql:"repositoryOwner(login: $login)"`
}

type User struct {
	Login             string
	Name              string
	Location          string
	Url               string
	Bio               string
	TwitterUsername   string
	IsViewer          bool
	IsFollowingViewer bool
	ViewerIsFollowing bool
	WebsiteUrl        string
	Followers         struct {
		TotalCount int
	}
	Following struct {
		TotalCount int
	}
	PinnedItems             PinnedItems     `graphql:"pinnedItems(first: $first, types: [REPOSITORY])"`
	Repositories            TopRepositories `graphql:"repositories(first: $first, privacy: PUBLIC, isFork: false, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC})"`
	ContributionsCollection struct {
		ContributionCalendar struct {
			TotalContributions int
			Weeks              []struct {
				ContributionDays []struct {
					ContributionLevel string
				}
			}
		}
	}
}

type Organization struct {
	Login           string
	Name            string
	Location        string
	Url             string
	Description     string
	TwitterUsername string
	WebsiteUrl      string
	PinnedItems     PinnedItems     `graphql:"pinnedItems(first: $first, types: [REPOSITORY])"`
	Repositories    TopRepositories `graphql:"repositories(first: $first, privacy: PUBLIC, isFork: false, ownerAffiliations: [OWNER], orderBy: {field: STARGAZERS, direction: DESC})"`
}

type PinnedItems struct {
	Nodes []PinnedItem
}

// PinnedItem is one pinned item; only repositories are queried.
type PinnedItem struct {
	Repository Repository `graphql:"... on Repository"`
}

// TopRepositories are an owner's public, non-fork repositories, most
// starred first: the fallback when nothing is pinned.
type TopRepositories struct {
	Nodes []Repository
}

type Repository struct {
	Id               string
	Name             string
	Description      string
	StargazerCount   int
	ViewerHasStarred bool
	Url              string
	Owner            struct {
		Login string
	}
	PrimaryLanguage struct {
		Name  string
		Color string
	}
}

// GetViewerQuery looks up the authenticated user's login.
type GetViewerQuery struct {
	Viewer struct {
		Login string
	}
}

// GetFollowersQuery pages through the people who follow a user.
type GetFollowersQuery struct {
	User struct {
		Followers People `graphql:"followers(first: $first, after: $after)"`
	} `graphql:"user(login: $login)"`
}

// GetFollowingQuery pages through the people a user follows.
type GetFollowingQuery struct {
	User struct {
		Following People `graphql:"following(first: $first, after: $after)"`
	} `graphql:"user(login: $login)"`
}

type People struct {
	TotalCount int
	PageInfo   struct {
		HasNextPage bool
		EndCursor   string
	}
	Nodes []struct {
		Login string
		Name  string
		Url   string
	}
}

// String and Int carry their GraphQL type names, which the client derives
// variable declarations from.
type (
	String string
	Int    int32
)
