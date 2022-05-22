package repo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/ghx/api"
	"github.com/coloradocolby/ghx/ui/styles"
)

func buildRepoDisplay(repo api.Repo, width int) string {
	// prep the finished matrix
	var b strings.Builder
	w := b.WriteString

	w(styles.Bold.Render(repo.Name) + "\n")
	w(fmt.Sprintf("%v %s", repo.Description, strings.Repeat(" ", width)))
	w("\n")
	coloredCircle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: repo.PrimaryLanguage.Color, Dark: repo.PrimaryLanguage.Color}).Render("●")
	w(fmt.Sprintf("%v %v  ⭑ %v", coloredCircle, repo.PrimaryLanguage.Name, repo.StargazerCount))

	return lipgloss.NewStyle().
		Align(lipgloss.Left).Render(styles.RepoFrame.Render(b.String()))
}

func BuildPinnedRepoDisplay(repos []struct{ Repo api.Repo }) string {
	var leftColB strings.Builder
	lw := leftColB.WriteString

	var rightColB strings.Builder
	rw := rightColB.WriteString

	maxLengthDesc := 0
	for _, r := range repos {
		if len(r.Repo.Description) > maxLengthDesc {
			maxLengthDesc = len(r.Repo.Description)
		}
	}

	for i, r := range repos {
		if i%2 == 0 {
			// left col
			lw(buildRepoDisplay(r.Repo, maxLengthDesc-len(r.Repo.Description)) + "\n\n")

		} else {
			rw(buildRepoDisplay(r.Repo, maxLengthDesc-len(r.Repo.Description)) + "\n\n")
			// right col
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColB.String(), rightColB.String())
}
