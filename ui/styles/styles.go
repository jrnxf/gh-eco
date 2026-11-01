// Package styles holds the colors and styles shared across the UI. Light
// colors meet WCAG AA against a white background: 4.5:1 for text and 3:1
// for borders, separators and other non-text marks. Dark colors are the
// ones gh-eco has always used.
package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Accent marks what has focus (5.05:1 on white).
	Accent = lipgloss.AdaptiveColor{Light: "#8250DF", Dark: "#7D56F4"}
	// Dim is for chrome that should recede, such as separators (3.45:1 on
	// white).
	Dim = lipgloss.AdaptiveColor{Light: "#818B98", Dark: "#484F58"}
	// CardBorder outlines unfocused cards: GitHub's light border, pale on
	// purpose (1.47:1 on white), so the focused card's border stands out.
	CardBorder = lipgloss.AdaptiveColor{Light: "#D0D7DE", Dark: "#484F58"}
	// mutedLight is secondary text on light backgrounds (6.11:1 on white).
	mutedLight = lipgloss.Color("#59636E")
	// Star is a starred repo's star (4.87:1 on white).
	Star = lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#DAAA3F"}
	// Spinner colors the loading spinner (5.19:1 on white).
	Spinner = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#5E81AC"}

	Bold = lipgloss.NewStyle().Bold(true)
	// FocusedName marks the profile name when it has focus: accent text,
	// underlined, so the underline is accent too.
	FocusedName = Bold.Underline(true).Foreground(Accent)

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

	// Frame outlines a card; a focused card's border turns accent.
	Frame        = lipgloss.NewStyle().Border(roundedBorder, true).Padding(0, 1).Margin(0, 1).BorderForeground(CardBorder)
	FocusedFrame = Frame.BorderForeground(Accent)

	// HelpKey, HelpDesc and HelpSeparator style the help bar. The dark
	// colors are bubbles' defaults; its light ones are too pale on white.
	HelpKey       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#424A53", Dark: "#626262"})
	HelpDesc      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#59636E", Dark: "#4A4A4A"})
	HelpSeparator = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#818B98", Dark: "#3C3C3C"})
)

// Muted is for secondary text. On a dark background it is faint, as
// before; on a light one faint text is too pale to read, so it is an
// explicit gray instead.
func Muted() lipgloss.Style {
	if lipgloss.HasDarkBackground() {
		return lipgloss.NewStyle().Faint(true)
	}
	return lipgloss.NewStyle().Foreground(mutedLight)
}
