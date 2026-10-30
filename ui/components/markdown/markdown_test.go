package markdown

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/ui/models"
)

const tableReadme = "# demo\n\n| a | b |\n|---|---|\n| 1 | 2 |\n"

func Test_ReadmeTableFitsViewport(t *testing.T) {
	ctx := &context.ProgramContext{}
	m := NewModel()
	m.UpdateProgramContext(ctx)

	ctx.Layout.ContentWidth, ctx.Layout.ContentHeight = 120, 40
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(commands.GetReadmeResponse{Readme: models.Blob{Text: tableReadme}})

	// shrinking after the readme loaded must re-wrap it to the new width
	ctx.Layout.ContentWidth = 80
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	if m.Viewport.Width != 80 {
		t.Fatalf("viewport width = %d, want 80", m.Viewport.Width)
	}
	// a stale wider render would be wrapped by the viewport onto extra
	// lines, so the visible lines must match a fresh 80-col render
	want := strings.Split(m.renderReadme(), "\n")
	got := strings.Split(m.Viewport.View(), "\n")
	if len(got) < len(want) {
		t.Fatalf("viewport shows %d lines, want at least %d", len(got), len(want))
	}
	for i, line := range want {
		if w := lipgloss.Width(line); w > 80 {
			t.Fatalf("readme line is %d cols wide in an 80-col viewport: %q", w, line)
		}
		if g, l := strings.TrimRight(got[i], " "), strings.TrimRight(line, " "); g != l {
			t.Fatalf("viewport line %d = %q, want %q", i, g, l)
		}
	}
}
