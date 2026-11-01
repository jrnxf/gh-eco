// Package text holds display-width aware string helpers.
package text

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Ellipsis marks text that was cut short.
const Ellipsis = "…"

// Ellipsize cuts s to at most width display cells, ending in a single "…"
// with no space before it. Styled text keeps its styling.
func Ellipsize(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	if width < 1 {
		return ""
	}
	cut := ansi.Truncate(s, width-1, "")
	if visible := ansi.Strip(cut); strings.HasSuffix(visible, " ") {
		cut = ansi.Truncate(s, lipgloss.Width(strings.TrimRight(visible, " ")), "")
	}
	return cut + Ellipsis
}
