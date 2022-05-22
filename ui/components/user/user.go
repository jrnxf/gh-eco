package user

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/ui/components/graph"
	"github.com/coloradocolby/gh-eco/ui/components/repo"
	"golang.org/x/term"
)

var (
	// colors
	subtleC    = lipgloss.Color("#768390")
	highlightC = lipgloss.Color("#81A1C1")
	subtle     = lipgloss.NewStyle().Foreground(subtleC).Render
	// highlight = lipgloss.NewStyle().Bold(true).Background(highlightC).Foreground(subtleC).PaddingLeft(1).PaddingRight(1).Render
	bold = lipgloss.NewStyle().Bold(true).Render
)

type Model struct {
	User    api.User
	display string
	err     error
}

func NewModel() Model {
	return Model{
		User: api.User{},
		err:  nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case api.SearchUserResponse:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.User = msg.User
			m.err = nil
		}
		m.buildDisplay()
	}

	return m, cmd
}

func (m Model) buildUserDisplay() string {
	u := m.User

	var b strings.Builder
	w := b.WriteString

	w(bold(u.Name) + "\n\n")

	w(subtle(u.Bio) + "\n\n")

	w(fmt.Sprintf("%v %v · %v %v\n\n", u.Followers.TotalCount, "followers", u.Following.TotalCount, "following"))

	w(fmt.Sprintf("%v  |  %v  |  @%v", u.Location, u.WebsiteUrl, u.TwitterUsername))

	return b.String()

}

func (m *Model) buildDisplay() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	var b strings.Builder
	w := b.WriteString

	u := m.User
	if m.err != nil {
		m.display = "No user found"
	} else {

		// w(styles.Focus.Copy().Align(lipgloss.Center).Render(m.buildUserDisplay()))
		w(m.buildUserDisplay())

		w("\n\n\n")

		w(fmt.Sprintf("%v contributions", u.ContributionsCollection.ContributionCalendar.TotalContributions))
		w("\n")

		w(lipgloss.NewStyle().
			Align(lipgloss.Left).
			Render(graph.BuildGraphDisplay(u.ContributionsCollection.ContributionCalendar.Weeks)))

		w("\n\n\n")

		w(lipgloss.NewStyle().
			Align(lipgloss.Center).Render(repo.BuildPinnedRepoDisplay([]struct{ Repo api.Repo }(u.PinnedItems.Nodes))))

		m.display = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(physicalWidth).Render(b.String())
	}

}

func (m Model) View() string {
	return m.display
}
