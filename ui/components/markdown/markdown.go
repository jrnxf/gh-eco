package markdown

import (
	"fmt"
	"strings"

	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"github.com/coloradocolby/gh-eco/utils"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Viewport   viewport.Model
	ctx        *context.ProgramContext
	mdRenderer *glamour.TermRenderer
}

func NewModel() Model {

	markdownRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)

	return Model{
		Viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		mdRenderer: markdownRenderer,
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
		out, _ := m.mdRenderer.Render(msg.Readme.Text)
		m.Viewport.SetYOffset(0) // scroll to top
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
	m.Viewport.Height = m.ctx.Layout.ContentHeight - lipgloss.Height(m.footerView())
}

func (m Model) View() string {
	return fmt.Sprintf("%s%s%s", m.Viewport.View(), utils.GetNewLines(1), m.footerView())
}

func (m Model) footerView() string {
	scrollPercentage := fmt.Sprintf(" %.f%%", m.Viewport.ScrollPercent()*100)
	line := strings.Repeat("─", utils.MaxInt(0, m.Viewport.Width-lipgloss.Width(scrollPercentage)))

	return styles.FaintBold.Render(lipgloss.JoinHorizontal(lipgloss.Center, line, scrollPercentage))
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx

}
