package repo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/styles"
	"github.com/jrnxf/gh-eco/utils"
)

func buildRepoDisplay(repo models.Repo, width int, isFocused bool, viewerHasStarred bool) string {
	var b strings.Builder
	w := b.WriteString

	if isFocused {
		w(styles.FocusedBold.Render(repo.Name))
	} else {
		w(styles.Bold.Render(repo.Name))
	}
	w(utils.GetNewLines(1))

	w(fmt.Sprintf("%s %s", utils.TruncateText(repo.Description, 60), strings.Repeat(" ", width)))
	w(utils.GetNewLines(1))

	coloredCircle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: repo.PrimaryLanguage.Color, Dark: repo.PrimaryLanguage.Color}).Render("●")

	star := "⭑"
	if viewerHasStarred {
		star = lipgloss.NewStyle().Foreground(lipgloss.Color("#DAAA3F")).Render("⭑")
	}
	w(fmt.Sprintf("%s %s  %s %v", coloredCircle, repo.PrimaryLanguage.Name, star, repo.StarsCount))

	return lipgloss.NewStyle().
		Align(lipgloss.Left).Render(styles.Frame.Render(b.String()))
}

// BuildPinnedRepoDisplay lays the repos out in two columns. focused indexes
// repos, or is -1 when no repo has focus.
func BuildPinnedRepoDisplay(repos []models.Repo, focused int) string {
	var lc strings.Builder // left col
	var rc strings.Builder // right col

	maxRepoDescLength := 0
	for _, r := range repos {
		currRepoDescLength := len(utils.TruncateText(r.Description, 60))
		if currRepoDescLength > maxRepoDescLength {
			maxRepoDescLength = currRepoDescLength
		}
	}

	for i, r := range repos {
		currRepoDescLength := len(utils.TruncateText(r.Description, 60))
		display := buildRepoDisplay(r, maxRepoDescLength-currRepoDescLength, i == focused, r.ViewerHasStarred) + utils.GetNewLines(1)
		if i%2 == 0 {
			lc.WriteString(display)
		} else {
			rc.WriteString(display)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, lc.String(), rc.String())
}
