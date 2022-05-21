package utils

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/ghx/api"
)

func ConvertWeeklyContributionsToGraph(weeklyContributions []api.WeeklyContribution) string {
	log.Println("start")

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
	var (
		none           = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D333B", Dark: "#2D333B"}).Render("■ ")
		firstQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0E4429", Dark: "#0E4429"}).Render("■ ")
		secondQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#006D32", Dark: "#006D32"}).Render("■ ")
		thirdQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#26A641", Dark: "#26A641"}).Render("■ ")
		fourthQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#39D353", Dark: "#39D353"}).Render("■ ")
	)

	var b strings.Builder

	for _, row := range slice {
		for _, cell := range row {
			switch cell {
			case "NONE":
				b.WriteString(none)
			case "FIRST_QUARTILE":
				b.WriteString(firstQuartile)
			case "SECOND_QUARTILE":
				b.WriteString(secondQuartile)
			case "THIRD_QUARTILE":
				b.WriteString(thirdQuartile)
			case "FOURTH_QUARTILE":
				b.WriteString(fourthQuartile)
			}
		}
		b.WriteString("\n")
	}
	log.Println("end")

	return b.String()
}
