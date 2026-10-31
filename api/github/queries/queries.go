package queries

type GetUserQuery struct {
	User struct {
		Id                string
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
		PinnedItems struct {
			Nodes []PinnedItem
		} `graphql:"pinnedItems(first: $first, types: [REPOSITORY])"`
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
	} `graphql:"user(login: $login)"`
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
