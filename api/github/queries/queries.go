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
	PinnedItems             PinnedItems `graphql:"pinnedItems(first: $first, types: [REPOSITORY])"`
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
	PinnedItems     PinnedItems `graphql:"pinnedItems(first: $first, types: [REPOSITORY])"`
}

type PinnedItems struct {
	Nodes []PinnedItem
}

// PinnedItem is one pinned item; only repositories are queried.
type PinnedItem struct {
	Repository struct {
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
	} `graphql:"... on Repository"`
}

// String and Int carry their GraphQL type names, which the client derives
// variable declarations from.
type (
	String string
	Int    int32
)
