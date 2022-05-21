package pager

import (
	"fmt"
	"log"
	"strings"

	"github.com/coloradocolby/ghx/ui/context"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type Model struct {
	content  string
	ready    bool
	Viewport viewport.Model
	ctx      *context.ProgramContext
}

func NewModel() Model {
	return Model{
		ready: false,
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

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.

			// TODO: (fix) 1 represents the height of the text input, figure out a cleaner way to do this
			m.Viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight-1)
			m.Viewport.YPosition = headerHeight
			m.Viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.Viewport.SetContent(m.content)
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.Viewport.YPosition = headerHeight + 1
		} else {
			m.Viewport.Width = msg.Width
			m.Viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.Viewport))
		}

	case tea.KeyMsg:
		if m.ctx.Mode == context.NormalMode {
			log.Println("listening")
			// only listen for keyboard events if in normal mode
			m.Viewport, cmd = m.Viewport.Update(msg)
		}
	default:
		// Handle other events (like mouse events)
		m.Viewport, cmd = m.Viewport.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.Viewport.View(), m.footerView())
}

func (m Model) headerView() string {
	line := strings.Repeat("─", m.Viewport.Width)

	return lipgloss.JoinHorizontal(lipgloss.Center, line)
}

func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.Viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
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
