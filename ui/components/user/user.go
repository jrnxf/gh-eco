package user

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/components/graph"
	"github.com/coloradocolby/gh-eco/ui/components/pager"
	"github.com/coloradocolby/gh-eco/ui/components/repo"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/models"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"golang.org/x/term"
)

type Model struct {
	User    models.User
	pager   pager.Model
	display string
	err     error
	ctx     *context.ProgramContext
}

func NewModel() Model {
	return Model{
		User: models.User{},
		err:  nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case commands.GetUserResponse:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.User = msg.User
			m.ctx.User = msg.User
			m.err = nil
		}
		m.buildDisplay()
	case commands.FocusChange:
		m.buildDisplay()
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) buildUserDisplay() string {
	u := m.User

	var b strings.Builder
	w := b.WriteString

	if u.Name == "" && u.Login != "" {
		if m.ctx.CurrentFocus.FocusedWidget.Type == "UserDisplay" {
			w(styles.FocusedBold.Render(u.Login) + "\n\n")
		} else {
			w(styles.Bold.Render(u.Login) + "\n\n")
		}
	}
	if u.Name != "" {
		if m.ctx.CurrentFocus.FocusedWidget.Type == "UserDisplay" {
			w(styles.FocusedBold.Render(u.Name) + "\n\n")
		} else {
			w(styles.Bold.Render(u.Name) + "\n\n")
		}
	}

	if u.Bio != "" {
		w(lipgloss.NewStyle().Faint(true).Width(80).Align(lipgloss.Center).Render(u.Bio) + "\n\n")
	}

	w(fmt.Sprintf("%v %v · %v %v\n", u.FollowersCount, "followers", u.FollowingCount, "following"))

	if (u.Location != "") || (u.WebsiteUrl != "") || (u.TwitterUsername != "") {
		line := []string{}
		if u.Location != "" {
			line = append(line, u.Location)
		}
		if u.WebsiteUrl != "" {
			line = append(line, u.WebsiteUrl)
		}
		if u.TwitterUsername != "" {
			line = append(line, fmt.Sprintf("@%s", u.TwitterUsername))
		}
		w("\n")
		w(strings.Join(line, "  |  "))
		w("\n")
	}

	m.ctx.FocusableWidgets = append(m.ctx.FocusableWidgets, context.FocusableWidget{Type: "UserDisplay", Info: struct {
		Url      string
		Owner    string
		RepoName string
	}{
		Url:      m.User.Url,
		Owner:    m.User.Login,
		RepoName: m.User.Login,
	}})

	return b.String()
}

func (m *Model) buildDisplay() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	var b strings.Builder
	w := b.WriteString

	u := m.User
	if m.err != nil {
		m.pager.Viewport.SetContent("No user found")
	} else {

		w(m.buildUserDisplay())

		w("\n\n")

		w(fmt.Sprintf("%v contributions", u.ActivityGraph.ContributionsCount))

		w("\n")

		w(lipgloss.NewStyle().
			Align(lipgloss.Left).
			Render(graph.BuildGraphDisplay(u.ActivityGraph.Weeks)))

		w("\n\n")

		w(lipgloss.NewStyle().
			Align(lipgloss.Center).Render(repo.BuildPinnedRepoDisplay(u.PinnedRepos, m.ctx)))

		m.display = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(physicalWidth).Render(b.String())
	}

}

func (m Model) View() string {
	return lipgloss.NewStyle().Height(m.ctx.Layout.ContentHeight).Render(m.display)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.pager.UpdateProgramContext(ctx)

	m.ctx = ctx
}
