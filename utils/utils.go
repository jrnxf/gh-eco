package utils

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"github.com/coloradocolby/gh-eco/api/github/queries"
	"github.com/coloradocolby/gh-eco/ui/models"
)

func TruncateText(str string, max int) string {
	if max <= 0 {
		return ""
	}

	lastSpaceIdx := -1
	len := 0
	for i, r := range str {
		if unicode.IsSpace(r) {
			lastSpaceIdx = i
		}
		len++
		if len > max {
			if lastSpaceIdx != -1 {
				return str[:lastSpaceIdx] + "..."
			}
			// string is longer than max but has no spaces
		}
	}
	// string is shorter than max
	return str
}

func BrowserOpen(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Println(err)
	}
}

func GetNewLines(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func MapGetUserQueryToDisplayUser(query queries.GetUserQuery) models.User {
	qu := query.User
	du := models.User{
		Login:           qu.Login,
		Name:            qu.Name,
		Location:        qu.Location,
		Url:             qu.Url,
		Bio:             qu.Bio,
		TwitterUsername: qu.TwitterUsername,
		WebsiteUrl:      qu.WebsiteUrl,
		FollowersCount:  qu.Followers.TotalCount,
		FollowingCount:  qu.Following.TotalCount,
	}

	du.ActivityGraph.ContributionsCount = qu.ContributionsCollection.ContributionCalendar.TotalContributions

	for _, week := range qu.ContributionsCollection.ContributionCalendar.Weeks {
		du.ActivityGraph.Weeks = append(du.ActivityGraph.Weeks, week)
	}

	for _, node := range qu.PinnedItems.Nodes {
		r := node.Repository
		du.PinnedRepos = append(du.PinnedRepos, models.Repo{
			Id:               r.Id,
			Name:             r.Name,
			Description:      r.Description,
			StarsCount:       r.StargazerCount,
			ViewerHasStarred: r.ViewerHasStarred,
			Owner: struct{ Login string }{
				Login: r.Owner.Login,
			},
			Url:             r.Url,
			PrimaryLanguage: r.PrimaryLanguage,
		})
	}

	return du
}
