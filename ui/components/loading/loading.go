// Package loading shows a spinner in the bottom row while a request is
// pending, but only once it has taken long enough to notice, so instant
// answers such as cached readmes never flicker.
package loading

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Delay is how long a request must be pending before the spinner shows.
const Delay = 150 * time.Millisecond

var labelStyle = lipgloss.NewStyle().Faint(true)

// ShowMsg reveals the spinner for request Req if it is still pending.
type ShowMsg struct {
	Req int
}

type Model struct {
	spinner spinner.Model
	delay   time.Duration
	// req identifies the latest Start; label is empty when nothing is pending.
	req     int
	label   string
	visible bool
}

func New(delay time.Duration) Model {
	return Model{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#5E81AC"))),
		),
		delay: delay,
	}
}

// Start marks a request labelled label as pending, replacing any other.
func (m *Model) Start(label string) tea.Cmd {
	m.req++
	m.label, m.visible = label, false
	req := m.req
	return tea.Tick(m.delay, func(time.Time) tea.Msg { return ShowMsg{Req: req} })
}

// Stop ends the pending request labelled label; another request that has
// replaced it keeps going.
func (m *Model) Stop(label string) {
	if m.label == label {
		m.label, m.visible = "", false
	}
}

// Visible reports whether the spinner is showing.
func (m Model) Visible() bool {
	return m.visible
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowMsg:
		if msg.Req == m.req && m.label != "" && !m.visible {
			m.visible = true
			return m, m.spinner.Tick
		}
	case spinner.TickMsg:
		if m.visible {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) View() string {
	return m.spinner.View() + " " + labelStyle.Render(m.label)
}
