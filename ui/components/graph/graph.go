package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/ui/models"
	"github.com/coloradocolby/gh-eco/utils"
)

var (
	GH_GRAPH_CELL                 = "■ "
	GH_GRAPH_CELL_NONE            = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#EBEDF0", Dark: "#2D333B"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_FIRST_QUARTILE  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9BE9A8", Dark: "#0E4429"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_SECOND_QUARTILE = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#40C463", Dark: "#006D32"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_THIRD_QUARTILE  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#30A14E", Dark: "#26A641"}).Render(GH_GRAPH_CELL)
	GH_GRAPH_CELL_FOURTH_QUARTILE = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#216E39", Dark: "#39D353"}).Render(GH_GRAPH_CELL)
)

func BuildGraphDisplay(weeklyContributions []models.WeeklyContribution) string {
	// prep the finished matrix

	result := make([][]string, len(weeklyContributions))
	for i := range result {
		result[i] = make([]string, len(weeklyContributions[0].ContributionDays))
	}

	for i, weeklyContribution := range weeklyContributions {
		for j, contributionDay := range weeklyContribution.ContributionDays {
			result[i][j] = contributionDay.ContributionLevel
		}
	}

	result = transposeSlice(result)

	foo := generateContributionGraph(result)

	return foo
}

func transposeSlice(slice [][]string) [][]string {
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

		w(utils.GetNewLines(1))
	}

	return b.String()
}
