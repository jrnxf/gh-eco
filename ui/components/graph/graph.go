package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/models"
)

const cell = "■"

// levels maps a contribution level to its color: GitHub's ramps for light
// and dark backgrounds.
var levels = map[string]lipgloss.AdaptiveColor{
	"NONE":            {Light: "#EBEDF0", Dark: "#2D333B"},
	"FIRST_QUARTILE":  {Light: "#9BE9A8", Dark: "#0E4429"},
	"SECOND_QUARTILE": {Light: "#40C463", Dark: "#006D32"},
	"THIRD_QUARTILE":  {Light: "#30A14E", Dark: "#26A641"},
	"FOURTH_QUARTILE": {Light: "#216E39", Dark: "#39D353"},
}

// cells renders one cell per level. It runs at render time rather than at
// package init, which happens before the terminal background is known.
func cells() map[string]string {
	out := make(map[string]string, len(levels))
	for level, color := range levels {
		out[level] = lipgloss.NewStyle().PaddingRight(1).Foreground(color).Render(cell)
	}
	return out
}

func BuildGraphDisplay(weeklyContributions []models.WeeklyContribution) string {
	if len(weeklyContributions) == 0 {
		return ""
	}

	// size rows by the longest week so ragged weeks can't index out of range;
	// unfilled cells stay "" and render as blanks
	maxDays := 0
	for _, week := range weeklyContributions {
		maxDays = max(maxDays, len(week.ContributionDays))
	}
	if maxDays == 0 {
		return ""
	}

	result := make([][]string, len(weeklyContributions))
	for i := range result {
		result[i] = make([]string, maxDays)
	}

	for i, weeklyContribution := range weeklyContributions {
		for j, contributionDay := range weeklyContribution.ContributionDays {
			result[i][j] = contributionDay.ContributionLevel
		}
	}

	return generateContributionGraph(transposeSlice(result))
}

func transposeSlice(slice [][]string) [][]string {
	if len(slice) == 0 {
		return slice
	}
	xLen := len(slice[0])
	yLen := len(slice)

	// prep the finished matrix
	result := make([][]string, xLen) // num empty rows to create (outer slice)
	for i := range result {
		result[i] = make([]string, yLen) // num empty columns to create in each row (inner slice)
	}

	for i := 0; i < xLen; i++ {
		for j := 0; j < yLen; j++ {
			result[i][j] = slice[j][i]
		}
	}
	return result
}

func generateContributionGraph(slice [][]string) string {
	var b strings.Builder
	rendered := cells()
	for _, row := range slice {
		for _, level := range row {
			b.WriteString(rendered[level])
		}
		b.WriteString("\n")
	}
	return b.String()
}
