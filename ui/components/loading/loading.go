// Package loading shows a spinner in the bottom row while a request is out
// on the network.
package loading

import (
	"slices"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/ui/styles"
)

type Model struct {
	spinner spinner.Model
	// labels names the pending requests, oldest first; the newest shows.
	labels []string
}

func New() Model {
	return Model{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(styles.Spinner)),
		),
	}
}

// Start shows the spinner for a request labelled label, on top of any
// other pending request.
func (m *Model) Start(label string) tea.Cmd {
	wasIdle := !m.Visible()
	m.Stop(label)
	m.labels = append(m.labels, label)
	if wasIdle {
		return m.spinner.Tick
	}
	// already ticking
	return nil
}

// Stop drops the request labelled label; the spinner keeps going while
// another request is pending.
func (m *Model) Stop(label string) {
	m.labels = slices.DeleteFunc(m.labels, func(l string) bool { return l == label })
}

// Visible reports whether the spinner is showing.
func (m Model) Visible() bool {
	return len(m.labels) > 0
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(spinner.TickMsg); ok && m.Visible() {
		// ticks stop once nothing is pending
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the spinner and label, cut to width.
func (m Model) View(width int) string {
	label := ""
	if m.Visible() {
		label = m.labels[len(m.labels)-1]
	}
	return ansi.Truncate(m.spinner.View()+" "+styles.Muted().Render(label), width, "…")
}
