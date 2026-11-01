// Package loading shows a spinner in the bottom row while a request is out
// on the network.
package loading

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/ui/styles"
)

type Model struct {
	spinner spinner.Model
	// label names the pending request; empty when nothing is pending.
	label string
}

func New() Model {
	return Model{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(styles.Spinner)),
		),
	}
}

// Start shows the spinner for a request labelled label, replacing any other.
func (m *Model) Start(label string) tea.Cmd {
	wasIdle := m.label == ""
	m.label = label
	if wasIdle {
		return m.spinner.Tick
	}
	// already ticking
	return nil
}

// Stop hides the spinner if it is showing for label; a request that has
// replaced it keeps going.
func (m *Model) Stop(label string) {
	if m.label == label {
		m.label = ""
	}
}

// Visible reports whether the spinner is showing.
func (m Model) Visible() bool {
	return m.label != ""
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(spinner.TickMsg); ok && m.label != "" {
		// ticks stop once nothing is pending
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the spinner and label, cut to width.
func (m Model) View(width int) string {
	return ansi.Truncate(m.spinner.View()+" "+styles.Muted().Render(m.label), width, "…")
}
