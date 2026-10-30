package markdown

import (
	"fmt"
	"os"
	"strings"

	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/ui/styles"
	"github.com/jrnxf/gh-eco/utils"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	glamourstyles "github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const maxWordWrap = 100

type Model struct {
	Viewport viewport.Model
	ctx      *context.ProgramContext
	style    ansi.StyleConfig
	readme   string
}

func NewModel() Model {
	return Model{
		Viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
		style: autoStyle(),
	}
}

// autoStyle mirrors glamour.WithAutoStyle but resolves the terminal
// background once, so the renderer can be rebuilt on resize without
// querying the terminal while bubbletea owns it.
func autoStyle() ansi.StyleConfig {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return glamourstyles.NoTTYStyleConfig
	}
	if lipgloss.HasDarkBackground() {
		return glamourstyles.DarkStyleConfig
	}
	return glamourstyles.LightStyleConfig
}

// renderReadme wraps to the viewport width (capped at maxWordWrap) because
// glamour stretches tables to the full word-wrap width.
func (m Model) renderReadme() string {
	wordWrap := maxWordWrap
	if m.Viewport.Width > 0 {
		wordWrap = min(m.Viewport.Width, maxWordWrap)
	}
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(m.style),
		glamour.WithWordWrap(wordWrap),
	)
	out, _ := renderer.Render(m.readme)
	return out
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
		m.readme = msg.Readme.Text
		m.Viewport.SetYOffset(0) // scroll to top
		m.Viewport.SetContent(m.renderReadme())

	case commands.LayoutChange:
		m.calculateViewportDimensions()

	case tea.WindowSizeMsg:
		m.calculateViewportDimensions()
		m.Viewport.SetContent(m.renderReadme())

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
