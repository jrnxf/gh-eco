package repo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/models"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"github.com/coloradocolby/gh-eco/utils"
)

func buildRepoDisplay(repo models.Repo, width int, isFocused bool, viewerHasStarred bool) string {
	// prep the finished matrix
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

func BuildPinnedRepoDisplay(repos []models.Repo, ctx *context.ProgramContext) string {
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

		widgetName := fmt.Sprintf("PinnedRepo%v", i+1)
		ctx.FocusableWidgets = append(ctx.FocusableWidgets, context.FocusableWidget{Type: context.RepoWidget, Repo: r, Descriptor: widgetName})
		fw := ctx.CurrentFocus.FocusedWidget
		display := buildRepoDisplay(r, maxRepoDescLength-currRepoDescLength, fw.Descriptor == widgetName, r.ViewerHasStarred) + utils.GetNewLines(1)
		if i%2 == 0 {
			lc.WriteString(display)
		} else {
			rc.WriteString(display)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, lc.String(), rc.String())
}
