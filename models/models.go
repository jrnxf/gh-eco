package models

type User struct {
	Id                string
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
	Readme          Blob
	PrimaryLanguage struct {
		Name  string
		Color string
	}
}

type Blob struct {
	Text string
}

type Starrable struct {
	Id               string
	StargazerCount   int
	ViewerHasStarred bool
}
