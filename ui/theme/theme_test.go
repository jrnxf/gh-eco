package theme

import "testing"

func Test_ParseOSC11(t *testing.T) {
	tests := []struct {
		reply string
		ok    bool
		dark  bool
	}{
		{"\x1b]11;rgb:ffff/ffff/ffff\x07", true, false},
		{"\x1b]11;rgb:0000/0000/0000\x1b\\", true, true},
		{"\x1b]11;rgb:fd/f6/e3\x07", true, false}, // solarized light, two digits
		{"\x1b]11;rgb:2828/2c2c/3434\x07", true, true},
		{"\x1b]11;rgb:f/f/f\x07", true, false},
		{"\x1b]11;?\x07", false, false},
		{"\x1b]11;rgb:zz/00/00\x07", false, false},
		{"", false, false},
	}
	for _, tt := range tests {
		r, g, b, ok := parseOSC11(tt.reply)
		if ok != tt.ok {
			t.Errorf("parseOSC11(%q) ok = %v, want %v", tt.reply, ok, tt.ok)
			continue
		}
		if ok && isDark(r, g, b) != tt.dark {
			t.Errorf("parseOSC11(%q) = %.2f,%.2f,%.2f, dark = %v, want %v", tt.reply, r, g, b, !tt.dark, tt.dark)
		}
	}
}
