// Package text holds display-width aware string helpers.
package text

import (
	"strconv"
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

// Count formats n with thousands separators, as GitHub does.
func Count(n int) string {
	if n < 0 {
		return "-" + Count(-n)
	}
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// Compact shortens a count for tight spots, GitHub style: 842, 1.2k, 21.8k,
// 1.5m. It rounds down so a count never reads higher than it is, and drops a
// trailing ".0".
func Compact(n int) string {
	switch {
	case n < 0:
		return "-" + Compact(-n)
	case n < 1_000:
		return strconv.Itoa(n)
	case n < 1_000_000:
		return tenths(n/100, "k")
	default:
		return tenths(n/100_000, "m")
	}
}

// tenths renders t tenths of a unit, such as 218 as "21.8".
func tenths(t int, suffix string) string {
	if t%10 == 0 {
		return strconv.Itoa(t/10) + suffix
	}
	return strconv.Itoa(t/10) + "." + strconv.Itoa(t%10) + suffix
}
