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
	// Repos are the pinned repositories or, when nothing is pinned and
	// TopRepos is set, the most-starred ones.
	Repos         []Repo
	TopRepos      bool
	ActivityGraph struct {
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

// PeopleKind picks a user's followers or the people they follow.
type PeopleKind int

const (
	Followers PeopleKind = iota
	Following
)

// Person is one entry of a followers or following list.
type Person struct {
	Login string
	Name  string
	Url   string
}

// PeoplePage is one page of a followers or following list. EndCursor
// continues the list when HasNextPage is set.
type PeoplePage struct {
	People      []Person
	TotalCount  int
	HasNextPage bool
	EndCursor   string
}
