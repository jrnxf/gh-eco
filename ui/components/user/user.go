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
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"golang.org/x/term"
)

type Model struct {
	User    api.User
	display string
	err     error
	ctx     *context.ProgramContext
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
	case context.FocusChange:
		// log.Println(m.ctx.CurrentFocus)
		m.buildDisplay()
	}

	return m, cmd
}

func (m Model) buildUserDisplay() string {
	u := m.User

	var b strings.Builder
	w := b.WriteString

	if m.ctx.CurrentFocus.FocusedWidget.Name == "UserDisplay" {
		w(styles.FocusedBold.Render(u.Name) + "\n\n")
	} else {
		w(styles.Bold.Render(u.Name) + "\n\n")
	}

	w(styles.Subtle.Render(u.Bio) + "\n\n")

	w(fmt.Sprintf("%v %v · %v %v\n\n", u.Followers.TotalCount, "followers", u.Following.TotalCount, "following"))

	w(fmt.Sprintf("%v  |  %v  |  @%v", u.Location, u.WebsiteUrl, u.TwitterUsername))

	m.ctx.FocusableWidgets = append(m.ctx.FocusableWidgets, context.FocusableWidget{Name: "UserDisplay"})
	// if m.ctx.CurrentFocus.FocusedWidget.Name == "UserDisplay" {
	// 	return styles.FocusedFrame.Copy().Align(lipgloss.Center).Render(b.String())

	// } else {
	// 	return styles.Frame.Copy().Align(lipgloss.Center).Render(b.String())
	// }
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
			Align(lipgloss.Center).Render(repo.BuildPinnedRepoDisplay([]struct{ Repo api.Repo }(u.PinnedItems.Nodes), m.ctx)))

		m.display = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(physicalWidth).Render(b.String())
	}

}

func (m Model) View() string {
	return m.display
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}
