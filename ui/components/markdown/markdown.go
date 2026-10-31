package markdown

import (
	"fmt"
	"os"
	"strings"

	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
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
	state    *state.State
	style    ansi.StyleConfig
	readme   string
}

func NewModel(s *state.State) Model {
	return Model{
		Viewport: viewport.New(0, 0),
		state:    s,
		style:    autoStyle(),
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state.Mode == state.NormalMode {
			// only listen for keyboard events if in normal mode
			m.Viewport, cmd = m.Viewport.Update(msg)
		}

	case commands.GetReadmeResponse:
		if msg.Err == nil {
			m.readme = msg.Readme
			m.Viewport.SetContent(m.renderReadme())
			m.Viewport.GotoTop()
		}

	default:
		// Handle other events (like mouse events)
		m.Viewport, cmd = m.Viewport.Update(msg)
	}

	m.fitViewport()
	return m, cmd
}

// fitViewport sizes the viewport to the content area, re-wrapping the
// readme when the width changes.
func (m *Model) fitViewport() {
	l := m.state.Layout
	m.Viewport.Height = max(0, l.ContentHeight-lipgloss.Height(m.footerView()))
	if m.Viewport.Width != l.ContentWidth {
		m.Viewport.Width = l.ContentWidth
		m.Viewport.SetContent(m.renderReadme())
	}
}

func (m Model) View() string {
	return fmt.Sprintf("%s%s%s", m.Viewport.View(), utils.GetNewLines(1), m.footerView())
}

func (m Model) footerView() string {
	scrollPercentage := fmt.Sprintf(" %.f%%", m.Viewport.ScrollPercent()*100)
	line := strings.Repeat("─", utils.MaxInt(0, m.Viewport.Width-lipgloss.Width(scrollPercentage)))

	return styles.FaintBold.Render(lipgloss.JoinHorizontal(lipgloss.Center, line, scrollPercentage))
}
