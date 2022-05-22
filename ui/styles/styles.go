package styles

import "github.com/charmbracelet/lipgloss"

var (
	subtleC    = lipgloss.AdaptiveColor{Light: "#768390", Dark: "#768390"}
	highlightC = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	Subtle = lipgloss.NewStyle().Foreground(subtleC)
	Bold   = lipgloss.NewStyle().Bold(true)
	Faint  = lipgloss.NewStyle().Faint(true)

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

	RepoFrame = lipgloss.NewStyle().Border(roundedBorder, true).Margin(0, 2).Padding(0, 1)
	// Focus     = lipgloss.NewStyle().Border(roundedBorder, true).BorderForeground(highlight).Padding(0, 1)
	Focus = lipgloss.NewStyle().Border(roundedBorder, true).Padding(0, 1)

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().Faint(true).Border(tabBorder, true).BorderForeground(highlightC).Padding(0, 1)

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	activeTab = tab.Copy().Faint(false).Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
)
