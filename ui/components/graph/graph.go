package graph

import (
	"strings"

	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/utils"
)

func BuildGraphDisplay(weeklyContributions []api.WeeklyContribution) string {
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

	return generateContributionGraph(result)
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

	for _, row := range slice {
		for _, cell := range row {
			switch cell {
			case "NONE":
				b.WriteString(graphCellNone.Render(utils.GH_GRAPH_CELL))
			case "FIRST_QUARTILE":
				b.WriteString(graphCellFirstQuartile.Render(utils.GH_GRAPH_CELL))
			case "SECOND_QUARTILE":
				b.WriteString(graphCellSecondQuartile.Render(utils.GH_GRAPH_CELL))
			case "THIRD_QUARTILE":
				b.WriteString(graphCellThirdQuartile.Render(utils.GH_GRAPH_CELL))
			case "FOURTH_QUARTILE":
				b.WriteString(graphCellFourthQuartile.Render(utils.GH_GRAPH_CELL))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
