package graph

import "github.com/charmbracelet/lipgloss"

var (
	graphCellNone           = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D333B", Dark: "#2D333B"}).Render("■ ")
	graphCellFirstQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0E4429", Dark: "#0E4429"}).Render("■ ")
	graphCellSecondQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#006D32", Dark: "#006D32"}).Render("■ ")
	graphCellThirdQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#26A641", Dark: "#26A641"}).Render("■ ")
	graphCellFourthQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#39D353", Dark: "#39D353"}).Render("■ ")
)
