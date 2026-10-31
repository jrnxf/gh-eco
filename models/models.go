package models

// User is a GitHub user or, when IsOrganization is set, an organization.
type User struct {
	IsOrganization    bool
	Login             string
	Name              string
	Location          string
	Url               string
	Bio               string
	TwitterUsername   string
	WebsiteUrl        string
	FollowersCount    int
	FollowingCount    int
	IsViewer          bool
	IsFollowingViewer bool
	ViewerIsFollowing bool
	PinnedRepos       []Repo
	ActivityGraph     struct {
		ContributionsCount int
		Weeks              []WeeklyContribution
	}
}

type WeeklyContribution struct {
	ContributionDays []struct {
		ContributionLevel string
	}
}

type Repo struct {
	Id               string
	Name             string
	Description      string
	StarsCount       int
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

type Starrable struct {
	Id               string
	StargazerCount   int
	ViewerHasStarred bool
}
