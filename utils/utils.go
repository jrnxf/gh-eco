package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"unicode"

	"github.com/charmbracelet/glamour"
	"github.com/coloradocolby/gh-eco/types/display"
	"github.com/coloradocolby/gh-eco/types/queries"
)

func TruncateText(str string, max int) string {
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

func JsonToMd(obj interface{}) string {
	val, _ := json.MarshalIndent(obj, "", "    ")
	in := "```json\n" + string(val) + "\n```"
	out, _ := glamour.Render(in, "dark")
	return out
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
		log.Fatal(err)
	}
}

func MapGetUserQueryToDisplayUser(query queries.GetUser) display.User {
	qu := query.User
	du := display.User{
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
		du.PinnedRepos = append(du.PinnedRepos, display.Repo{
			Id:          r.Id,
			Name:        r.Name,
			Description: r.Description,
			StarsCount:  r.StargazerCount,
			Owner: struct{ Login string }{
				Login: r.Owner.Login,
			},
			Url:             r.Url,
			PrimaryLanguage: r.PrimaryLanguage,
		})
	}

	return du
}
