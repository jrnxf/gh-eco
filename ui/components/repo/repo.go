package repo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/styles"
)

// maxDescription is how many characters of a description a card shows
// before cutting at a word boundary.
const maxDescription = 60

// minTwoColumnText is the narrowest card text that still gets two columns.
const minTwoColumnText = 36

// cardChrome is the width a card adds around its text: border, padding and
// margin on both sides.
const cardChrome = 6

// Grid lays the repos out as cards, in two columns when width allows and
// one otherwise. focused indexes repos, or is -1 when no repo has focus.
// Every row of cards is rowHeight lines tall; top and bottom are the first
// and last lines of the focused card, or -1.
func Grid(repos []models.Repo, focused, width int) (grid string, rowHeight, top, bottom int) {
	if len(repos) == 0 {
		return "", 0, -1, -1
	}

	text := 0
	for _, r := range repos {
		text = max(text, lipgloss.Width(r.Name), lipgloss.Width(description(r)), lipgloss.Width(stats(r)))
	}
	// two columns while each card still has room for a readable
	// description, one otherwise
	columns := 2
	if half := width/2 - cardChrome; half < text {
		if half >= minTwoColumnText {
			text = half
		} else {
			columns = 1
			text = min(text, width-cardChrome)
		}
	}
	text = max(text, 1)

	cols := make([][]string, columns)
	for i, r := range repos {
		cols[i%columns] = append(cols[i%columns], card(r, text, i == focused))
	}
	rendered := make([]string, columns)
	for i, c := range cols {
		rendered[i] = lipgloss.JoinVertical(lipgloss.Left, c...)
	}
	grid = lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	rowHeight = lipgloss.Height(cols[0][0])
	if focused < 0 || focused >= len(repos) {
		return grid, rowHeight, -1, -1
	}
	top = focused / columns * rowHeight
	return grid, rowHeight, top, top + rowHeight - 1
}

// card renders one repo in a frame whose text area is width cells wide.
func card(r models.Repo, width int, focused bool) string {
	name := styles.Bold
	if focused {
		name = styles.FocusedBold
	}
	lines := []string{
		name.Render(ansi.Truncate(r.Name, width, "...")),
		ansi.Truncate(description(r), width, "..."),
		ansi.Truncate(stats(r), width, "..."),
	}
	return styles.Frame.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func description(r models.Repo) string {
	return truncateText(r.Description, maxDescription)
}

// stats is the language and star count line; repos without a detected
// language show only the stars.
func stats(r models.Repo) string {
	star := "⭑"
	if r.ViewerHasStarred {
		star = lipgloss.NewStyle().Foreground(lipgloss.Color("#DAAA3F")).Render("⭑")
	}
	stars := fmt.Sprintf("%s %v", star, r.StarsCount)
	lang := r.PrimaryLanguage
	if lang.Name == "" {
		return stars
	}
	dot := lipgloss.NewStyle().Foreground(lipgloss.Color(lang.Color)).Render("●")
	return fmt.Sprintf("%s %s  %s", dot, lang.Name, stars)
}
