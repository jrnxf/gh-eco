package styles

import "github.com/charmbracelet/lipgloss"

var (
	Bold        = lipgloss.NewStyle().Bold(true)
	FocusedBold = Bold.Copy().Background(lipgloss.AdaptiveColor{Light: "#5E81AC", Dark: "#5E81AC"})

	Faint     = lipgloss.NewStyle().Faint(true)
	FaintBold = Faint.Copy().Bold(true)

	roundedBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	Frame = lipgloss.NewStyle().Border(roundedBorder, true).Padding(0, 1).Margin(0, 1)
)
