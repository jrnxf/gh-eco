package repo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/styles"
)

func buildRepoDisplay(repo api.Repo, width int, isFocused bool) string {
	// prep the finished matrix
	var b strings.Builder
	w := b.WriteString

	if isFocused {
		w(styles.FocusedBold.Render(repo.Name))
	} else {
		w(styles.Bold.Render(repo.Name))
	}
	w("\n")

	w(fmt.Sprintf("%v %s", repo.Description, strings.Repeat(" ", width)))
	w("\n")
	coloredCircle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: repo.PrimaryLanguage.Color, Dark: repo.PrimaryLanguage.Color}).Render("●")
	w(fmt.Sprintf("%v %v  ⭑ %v", coloredCircle, repo.PrimaryLanguage.Name, repo.StargazerCount))

	return lipgloss.NewStyle().
		Align(lipgloss.Left).Render(styles.Frame.Render(b.String()))

}

func BuildPinnedRepoDisplay(repos []struct{ Repo api.Repo }, ctx *context.ProgramContext) string {
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
		widgetName := fmt.Sprintf("PinnedRepo%v", i+1)
		ctx.FocusableWidgets = append(ctx.FocusableWidgets, context.FocusableWidget{Name: widgetName})
		d := buildRepoDisplay(r.Repo, maxLengthDesc-len(r.Repo.Description), ctx.CurrentFocus.FocusedWidget.Name == widgetName) + "\n\n"
		if i%2 == 0 {
			// left col
			lw(d)
		} else {
			rw(d)
			// right col
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColB.String(), rightColB.String())
}
