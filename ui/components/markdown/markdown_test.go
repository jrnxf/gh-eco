package markdown

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

const tableReadme = "# demo\n\n| a | b |\n|---|---|\n| 1 | 2 |\n"

func Test_ReadmeTableFitsViewport(t *testing.T) {
	s := &state.State{}
	m := NewModel(s)

	s.Layout.ContentWidth, s.Layout.ContentHeight = 120, 40
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(commands.GetReadmeResponse{Readme: tableReadme})

	// shrinking after the readme loaded must re-wrap it to the new width
	s.Layout.ContentWidth = 80
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	// the viewport keeps state.RightSlack cells free at the right edge
	if want := 80 - state.RightSlack; m.Viewport.Width != want {
		t.Fatalf("viewport width = %d, want %d", m.Viewport.Width, want)
	}
	// a stale wider render would be wrapped by the viewport onto extra
	// lines, so the visible lines must match a fresh 80-col render
	want := strings.Split(m.renderReadme(), "\n")
	got := strings.Split(m.Viewport.View(), "\n")
	if len(got) < len(want) {
		t.Fatalf("viewport shows %d lines, want at least %d", len(got), len(want))
	}
	for i, line := range want {
		if w := lipgloss.Width(line); w > m.Viewport.Width {
			t.Fatalf("readme line is %d cols wide in a %d-col viewport: %q", w, m.Viewport.Width, line)
		}
		if g, l := strings.TrimRight(got[i], " "), strings.TrimRight(line, " "); g != l {
			t.Fatalf("viewport line %d = %q, want %q", i, g, l)
		}
	}
}
