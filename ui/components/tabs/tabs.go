package tabs

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/coloradocolby/ghx/ui/constants"
	"github.com/coloradocolby/ghx/ui/context"
	"golang.org/x/term"
)

type Model struct {
	CurrTab int
}

func NewModel() Model {
	return Model{
		CurrTab: 0,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(ctx context.ProgramContext) string {
	var tabs []string

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	for i, t := range constants.Tabs {
		if m.CurrTab == i {
			tabs = append(tabs, activeTab.Render(t))
		} else {
			tabs = append(tabs, tab.Render(t))
		}
	}

	tabRowTabX := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	tabRowGapX := tabGap.Render(strings.Repeat(" ", max(0, physicalWidth-lipgloss.Width(tabRowTabX)-2)))

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRowTabX, tabRowGapX)
	return row
}

func (m *Model) PrevTab() {
	if m.CurrTab == 0 {
		m.CurrTab = len(constants.Tabs) - 1
	} else {
		m.CurrTab -= 1
	}
}

func (m *Model) NextTab() {
	if m.CurrTab == len(constants.Tabs)-1 {
		m.CurrTab = 0
	} else {
		m.CurrTab += 1
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
