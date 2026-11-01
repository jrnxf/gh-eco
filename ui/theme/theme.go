// Package theme works out whether the terminal background is dark before
// the UI starts, so adaptive colors pick the right palette.
package theme

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Detect settles the background lipgloss assumes, logging the evidence to
// the debug log. Without an override it asks the terminal itself in every
// terminal: termenv refuses to inside tmux or screen and then assumes dark.
func Detect() {
	override := os.Getenv("GH_ECO_THEME")
	reply := ""
	if _, set := parseOverride(override); !set {
		var ok bool
		if reply, ok = queryBackground(); !ok {
			// not a terminal: lipgloss renders without color anyway
			return
		}
	}
	dark, why := decide(override, reply, os.Getenv("COLORFGBG"))
	log.Printf("background: dark=%v (%s)", dark, why)
	lipgloss.SetHasDarkBackground(dark)
}

// decide picks dark or light from, in order: GH_ECO_THEME (light or dark),
// the terminal's OSC 11 reply, the COLORFGBG variable some terminals set,
// and a dark default. The override exists for terminals that misreport
// their background, such as herdr 0.9.3, which answers OSC 11 with a dark
// color in a light pane.
func decide(override, reply, colorFGBG string) (dark bool, why string) {
	if dark, set := parseOverride(override); set {
		return dark, "GH_ECO_THEME=" + override
	}
	if r, g, b, ok := parseOSC11(reply); ok {
		return isDark(r, g, b), fmt.Sprintf("OSC 11 reply rgb %.2f/%.2f/%.2f", r, g, b)
	}
	if bg, ok := parseColorFGBG(colorFGBG); ok {
		// ANSI 7 (light gray) and the bright colors from 9 up are light
		return bg < 7 || bg == 8, fmt.Sprintf("COLORFGBG background %d", bg)
	}
	return true, fmt.Sprintf("no OSC 11 reply (%q) or COLORFGBG, default", reply)
}

// isDark matches termenv: a background is dark when its HSL lightness is
// under one half. r, g and b are 0-1.
func isDark(r, g, b float64) bool {
	return (max(r, g, b)+min(r, g, b))/2 < 0.5
}

// parseOSC11 finds a reply like "\x1b]11;rgb:ffff/ffff/ffff\x07" in s,
// where each channel has one to four hex digits.
func parseOSC11(s string) (r, g, b float64, ok bool) {
	i := strings.Index(s, "]11;rgb:")
	if i < 0 {
		return 0, 0, 0, false
	}
	body := s[i+len("]11;rgb:"):]
	if end := strings.IndexAny(body, "\x07\x1b"); end >= 0 {
		body = body[:end]
	}
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

// parseOverride reads GH_ECO_THEME: "light" or "dark", anything else means
// detect.
func parseOverride(v string) (dark, set bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "light":
		return false, true
	case "dark":
		return true, true
	}
	return false, false
}

// parseColorFGBG reads the background index from "fg;bg" or
// "fg;default;bg".
func parseColorFGBG(v string) (int, bool) {
	parts := strings.Split(v, ";")
	if len(parts) < 2 {
		return 0, false
	}
	bg, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || bg < 0 || bg > 15 {
		return 0, false
	}
	return bg, true
}
