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
	// check the rendered readme itself: the viewport would hide overflow
	// by wrapping it onto extra lines
	for _, line := range strings.Split(m.renderReadme(), "\n") {
		if w := lipgloss.Width(line); w > 80 {
			t.Fatalf("readme line is %d cols wide in an 80-col viewport: %q", w, line)
		}
	}
}
