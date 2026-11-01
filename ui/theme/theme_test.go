package theme

import (
	"strings"
	"testing"
)

func Test_DecideBackground(t *testing.T) {
	tests := []struct {
		name      string
		override  string
		reply     string
		colorFGBG string
		dark      bool
		source    string
	}{
		{"override light beats a dark OSC 11 reply", "light", "\x1b]11;rgb:1414/1414/1414\x07", "15;0", false, "GH_ECO_THEME"},
		{"override dark beats a light OSC 11 reply", "DARK", "\x1b]11;rgb:ffff/ffff/ffff\x07", "0;15", true, "GH_ECO_THEME"},
		{"unknown override means detect", "auto", "\x1b]11;rgb:ffff/ffff/ffff\x07", "", false, "OSC 11"},
		{"light OSC 11 reply", "", "\x1b]11;rgb:ffff/ffff/ffff\x07\x1b[?62;22c", "", false, "OSC 11"},
		{"dark OSC 11 reply, ST-terminated", "", "\x1b]11;rgb:1e1e/1e1e/2e2e\x1b\\\x1b[?1;2c", "", true, "OSC 11"},
		{"two-digit channels", "", "\x1b]11;rgb:fd/f6/e3\x07", "", false, "OSC 11"},
		{"OSC 11 beats COLORFGBG", "", "\x1b]11;rgb:ffff/ffff/ffff\x07", "15;0", false, "OSC 11"},
		{"no OSC 11, light COLORFGBG", "", "\x1b[?62;22c", "0;15", false, "COLORFGBG"},
		{"no OSC 11, three-part COLORFGBG", "", "\x1b[?62;22c", "0;default;7", false, "COLORFGBG"},
		{"no OSC 11, dark COLORFGBG", "", "", "15;0", true, "COLORFGBG"},
		{"garbled reply falls back", "", "\x1b]11;rgb:zz/00/00\x07", "0;15", false, "COLORFGBG"},
		{"nothing at all is dark", "", "\x1b[?62;22c", "", true, "default"},
		{"bad COLORFGBG is ignored", "", "", "default", true, "default"},
	}
	for _, tt := range tests {
		dark, why := decide(tt.override, tt.reply, tt.colorFGBG)
		if dark != tt.dark || !strings.Contains(why, tt.source) {
			t.Errorf("%s: dark = %v (%s), want %v from %s", tt.name, dark, why, tt.dark, tt.source)
		}
	}
}

func Test_AnsweredDA1(t *testing.T) {
	if answeredDA1("\x1b]11;rgb:ffff/ffff/ffff\x07") {
		t.Errorf("an OSC 11 reply alone is not the DA1 reply")
	}
	if !answeredDA1("\x1b]11;rgb:ffff/ffff/ffff\x07\x1b[?62;22c") {
		t.Errorf("missed the DA1 reply")
	}
}
