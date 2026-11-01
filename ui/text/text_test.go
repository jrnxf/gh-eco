package text

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func Test_Ellipsize(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"with results and historical logging", 27, "with results and historica…"},
		{"with results and historical logging", 28, "with results and historical…"},
		// a cut right after a space must not leave "word …"
		{"with results and historical logging", 29, "with results and historical…"},
		{"漢字の説明文", 7, "漢字の…"}, // wide runes count two cells
		{"🤘🏼 all things", 6, "🤘🏼 al…"},
		{"anything", 0, ""},
	}
	for _, tt := range tests {
		got := Ellipsize(tt.in, tt.width)
		if got != tt.want {
			t.Errorf("Ellipsize(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
		}
		if w := lipgloss.Width(got); w > tt.width {
			t.Errorf("Ellipsize(%q, %d) is %d cells wide", tt.in, tt.width, w)
		}
	}

	// styling survives the cut
	got := Ellipsize("\x1b[1mbold words here\x1b[0m", 10)
	if ansi.Strip(got) != "bold word…" || !strings.HasPrefix(got, "\x1b[1m") {
		t.Errorf("styled cut = %q", got)
	}
}
