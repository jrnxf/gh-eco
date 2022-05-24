package graph

import "github.com/charmbracelet/lipgloss"

var (
	graphCellNone           = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D333B", Dark: "#2D333B"})
	graphCellFirstQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0E4429", Dark: "#0E4429"})
	graphCellSecondQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#006D32", Dark: "#006D32"})
	graphCellThirdQuartile  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#26A641", Dark: "#26A641"})
	graphCellFourthQuartile = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#39D353", Dark: "#39D353"})
)
