// Package loading animates the spinner shown while a request is out on the
// network.
package loading

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jrnxf/gh-eco/ui/styles"
)

type Model struct {
	spinner spinner.Model
	// pending names the requests out on the network.
	pending []string
}

func New() Model {
	return Model{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(styles.Spinner)),
		),
	}
}

// Start marks the request named name as pending, starting the spinner if
// nothing else was.
func (m *Model) Start(name string) tea.Cmd {
	wasIdle := !m.Visible()
	m.Stop(name)
	m.pending = append(m.pending, name)
	if wasIdle {
		return m.spinner.Tick
	}
	// already ticking
	return nil
}

// Stop drops the request named name; the spinner keeps going while
// another request is pending.
func (m *Model) Stop(name string) {
	m.pending = slices.DeleteFunc(m.pending, func(n string) bool { return n == name })
}

// Pending reports whether the request named name is out.
func (m Model) Pending(name string) bool {
	return slices.Contains(m.pending, name)
}

// Visible reports whether any request is out.
func (m Model) Visible() bool {
	return len(m.pending) > 0
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

// View renders the spinner and label. Each frame of the dot spinner ends
// in a space, which separates the two.
func (m Model) View(label string) string {
	return m.spinner.View() + styles.Muted().Render(label)
}

// Spinner renders the spinner alone, one cell wide.
func (m Model) Spinner() string {
	return strings.TrimRight(m.spinner.View(), " ")
}
