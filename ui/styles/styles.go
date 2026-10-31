package styles

import "github.com/charmbracelet/lipgloss"

var (
	Bold        = lipgloss.NewStyle().Bold(true)
	FocusedBold = Bold.Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"})

	Faint     = lipgloss.NewStyle().Faint(true)
	FaintBold = Faint.Bold(true)

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
