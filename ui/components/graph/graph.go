package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/models"
)

var (
	GH_GRAPH_CELL                 = "■"
	GH_GRAPH_CELL_NONE            = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#EBEDF0", Dark: "#2D333B"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_FIRST_QUARTILE  = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#9BE9A8", Dark: "#0E4429"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_SECOND_QUARTILE = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#40C463", Dark: "#006D32"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_THIRD_QUARTILE  = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#30A14E", Dark: "#26A641"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_FOURTH_QUARTILE = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#216E39", Dark: "#39D353"}).Render(GH_GRAPH_CELL)
)

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
	w := b.WriteString

	for _, row := range slice {
		for _, cell := range row {
			switch cell {
			case "NONE":
				w(GH_GRAPH_CELL_NONE)
			case "FIRST_QUARTILE":
				w(GH_GRAPH_CELL_FIRST_QUARTILE)
			case "SECOND_QUARTILE":
				w(GH_GRAPH_CELL_SECOND_QUARTILE)
			case "THIRD_QUARTILE":
				w(GH_GRAPH_CELL_THIRD_QUARTILE)
			case "FOURTH_QUARTILE":
				w(GH_GRAPH_CELL_FOURTH_QUARTILE)
			}
		}

		w("\n")
	}

	return b.String()
}
