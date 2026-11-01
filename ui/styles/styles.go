package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Accent marks what has focus.
	Accent = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	// Dim is for chrome that should recede, such as unfocused borders.
	Dim = lipgloss.AdaptiveColor{Light: "#AFB8C1", Dark: "#484F58"}

	Bold        = lipgloss.NewStyle().Bold(true)
	FocusedBold = Bold.Background(Accent)

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

	Frame        = lipgloss.NewStyle().Border(roundedBorder, true).Padding(0, 1).Margin(0, 1).BorderForeground(Dim)
	FocusedFrame = Frame.BorderForeground(Accent)
)
