package queries

type GetUser struct {
	User struct {
		Login           string
		Name            string
		Location        string
		Url             string
		Bio             string
		TwitterUsername string
		WebsiteUrl      string
		Followers       struct {
			TotalCount int
		}
		Following struct {
			TotalCount int
		}
		PinnedItems struct {
			Nodes []struct {
				Repository struct {
					Id             string
					Name           string
					Description    string
					StargazerCount int
					Url            string
					Owner          struct {
						Login string
					}
					PrimaryLanguage struct {
						Name  string
						Color string
					}
				} `graphql:"... on Repository"`
			}
		} `graphql:"pinnedItems(first: $first)"`
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

type GetReadme struct {
	Repository struct {
		Object struct {
			Blob struct {
				Text string
			} `graphql:"... on Blob"`
		} `graphql:"object(expression: $expression)"`
	} `graphql:"repository(name: $name, owner: $owner)"`
}
