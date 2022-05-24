package styles

import "github.com/charmbracelet/lipgloss"

var (
	SubtleC    = lipgloss.AdaptiveColor{Light: "#768390", Dark: "#768390"}
	HighlightC = lipgloss.AdaptiveColor{Light: "#5E81AC", Dark: "#5E81AC"}

	Subtle      = lipgloss.NewStyle().Foreground(SubtleC)
	Bold        = lipgloss.NewStyle().Bold(true)
	FocusedBold = Bold.Copy().Background(HighlightC)
	Faint       = lipgloss.NewStyle().Faint(true)

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

	Frame = lipgloss.NewStyle().Border(roundedBorder, true).Margin(0, 2).Padding(0, 1)
)
