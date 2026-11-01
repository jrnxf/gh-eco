// Package theme works out whether the terminal background is dark before
// the UI starts, so adaptive colors pick the right palette.
package theme

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Detect settles the background lipgloss assumes. termenv refuses to ask
// the terminal inside tmux or screen and then assumes dark, which gives a
// light terminal the dark palette. tmux 3.3 and later answers the query
// itself with the client terminal's colors, so ask directly there.
func Detect() {
	if os.Getenv("COLORFGBG") != "" {
		// termenv reads it
		return
	}
	term := os.Getenv("TERM")
	if os.Getenv("TMUX") == "" && !strings.HasPrefix(term, "screen") && !strings.HasPrefix(term, "tmux") {
		// outside a multiplexer termenv's own query works
		return
	}
	if r, g, b, ok := queryBackground(); ok {
		lipgloss.SetHasDarkBackground(isDark(r, g, b))
	}
}

// isDark matches termenv: a background is dark when its HSL lightness is
// under one half. r, g and b are 0-1.
func isDark(r, g, b float64) bool {
	return (max(r, g, b)+min(r, g, b))/2 < 0.5
}

// parseOSC11 reads a reply like "\x1b]11;rgb:ffff/ffff/ffff\x07", where
// each channel has one to four hex digits.
func parseOSC11(reply string) (r, g, b float64, ok bool) {
	i := strings.Index(reply, "rgb:")
	if i < 0 {
		return 0, 0, 0, false
	}
	body := strings.TrimRight(reply[i+len("rgb:"):], "\x07\x1b\\")
	parts := strings.Split(body, "/")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	var ch [3]float64
	for k, p := range parts {
		if len(p) == 0 || len(p) > 4 {
			return 0, 0, 0, false
		}
		v, err := strconv.ParseUint(p, 16, 16)
		if err != nil {
			return 0, 0, 0, false
		}
		ch[k] = float64(v) / float64(uint64(1)<<(4*len(p))-1)
	}
	return ch[0], ch[1], ch[2], true
}
