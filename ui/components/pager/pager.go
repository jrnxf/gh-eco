package pager

import (
	"fmt"
	"strings"

	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Viewport viewport.Model
	ctx      *context.ProgramContext
}

func NewModel() Model {
	return Model{
		Viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.ctx.Mode == context.NormalMode {
			// only listen for keyboard events if in normal mode
			m.Viewport, cmd = m.Viewport.Update(msg)
		}

	case commands.GetReadmeResponse:
		out, _ := glamour.Render(msg.Readme.Text, "dark")
		m.Viewport.SetContent(out)

	case commands.LayoutChange:
		m.calculateViewportDimensions()

	case tea.WindowSizeMsg:
		m.calculateViewportDimensions()

	default:
		// Handle other events (like mouse events)
		m.Viewport, cmd = m.Viewport.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) calculateViewportDimensions() {
	m.Viewport.Width = m.ctx.Layout.ContentWidth
	m.Viewport.Height = m.ctx.Layout.ContentHeight - lipgloss.Height(m.headerView()) - lipgloss.Height(m.footerView())
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.Viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	line := strings.Repeat("─", m.Viewport.Width)

	return styles.FaintBold.Render(line)
}

func (m Model) footerView() string {
	scrollPercentage := fmt.Sprintf(" %.f%%", m.Viewport.ScrollPercent()*100)
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(scrollPercentage)))

	return styles.FaintBold.Render(lipgloss.JoinHorizontal(lipgloss.Center, line, scrollPercentage))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx

}
