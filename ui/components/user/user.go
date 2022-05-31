package user

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/components/graph"
	"github.com/coloradocolby/gh-eco/ui/components/repo"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/ui/models"
	"github.com/coloradocolby/gh-eco/ui/styles"
	"github.com/coloradocolby/gh-eco/utils"
	"golang.org/x/term"
)

type Model struct {
	User    models.User
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

	case tea.WindowSizeMsg:
		m.buildDisplay()

	case commands.FollowUserResponse:
		u := &m.ctx.User
		fw := &m.ctx.CurrentFocus.FocusedWidget
		u.FollowersCount = msg.User.FollowersCount
		u.ViewerIsFollowing = msg.User.ViewerIsFollowing
		fw.User.FollowersCount = msg.User.FollowersCount
		fw.User.ViewerIsFollowing = msg.User.ViewerIsFollowing
		m.buildDisplay()

	case commands.UnfollowUserResponse:
		u := &m.ctx.User
		fw := &m.ctx.CurrentFocus.FocusedWidget
		u.FollowersCount = msg.User.FollowersCount
		u.ViewerIsFollowing = msg.User.ViewerIsFollowing
		fw.User.FollowersCount = msg.User.FollowersCount
		fw.User.ViewerIsFollowing = msg.User.ViewerIsFollowing
		m.buildDisplay()

	case commands.StarStarrableResponse:
		for i, r := range m.ctx.User.PinnedRepos {
			if r.Id == msg.Starrable.Id {
				m.ctx.User.PinnedRepos[i].ViewerHasStarred = msg.Starrable.ViewerHasStarred
				m.ctx.User.PinnedRepos[i].StarsCount = msg.Starrable.StargazerCount
				m.ctx.CurrentFocus.FocusedWidget.Repo.ViewerHasStarred = msg.Starrable.ViewerHasStarred
				m.ctx.CurrentFocus.FocusedWidget.Repo.StarsCount = msg.Starrable.StargazerCount
			}
		}
		m.buildDisplay()

	case commands.RemoveStarStarrableResponse:
		for i, r := range m.ctx.User.PinnedRepos {
			if r.Id == msg.Starrable.Id {
				m.ctx.User.PinnedRepos[i].ViewerHasStarred = msg.Starrable.ViewerHasStarred
				m.ctx.User.PinnedRepos[i].StarsCount = msg.Starrable.StargazerCount
				m.ctx.CurrentFocus.FocusedWidget.Repo.ViewerHasStarred = msg.Starrable.ViewerHasStarred
				m.ctx.CurrentFocus.FocusedWidget.Repo.StarsCount = msg.Starrable.StargazerCount
			}
		}
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
		// user hasn't specified their name on github
		if m.ctx.CurrentFocus.FocusedWidget.Descriptor == "UserDisplay" {
			w(styles.FocusedBold.Render(u.Login))
		} else {
			w(styles.Bold.Render(u.Login))

		}
		w(utils.GetNewLines(2))

	}
	if u.Name != "" {
		if m.ctx.CurrentFocus.FocusedWidget.Descriptor == "UserDisplay" {
			w(styles.FocusedBold.Render(u.Name))
		} else {
			w(styles.Bold.Render(u.Name))
		}
		w(utils.GetNewLines(2))

	}

	if u.Bio != "" {
		w(lipgloss.NewStyle().Faint(true).Align(lipgloss.Center).Render(u.Bio))
		w(utils.GetNewLines(2))
	}

	var (
		viewerIsFollowingStr string
		isFollowingViewerStr string
	)

	if u.ViewerIsFollowing {
		viewerIsFollowingStr = lipgloss.NewStyle().Italic(true).Render(" (you follow)")
	}

	if u.IsFollowingViewer {
		isFollowingViewerStr = lipgloss.NewStyle().Italic(true).Render(" (follows you)")
	}

	w(fmt.Sprintf("%v %s%s / %v %s%s", u.FollowersCount, "followers", viewerIsFollowingStr, u.FollowingCount, "following", isFollowingViewerStr))
	w(utils.GetNewLines(1))

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

		w(utils.GetNewLines(1))
		w(strings.Join(line, "  /  "))
		w(utils.GetNewLines(1))
	}

	m.ctx.FocusableWidgets = append(m.ctx.FocusableWidgets, context.FocusableWidget{Descriptor: "UserDisplay", Type: context.UserWidget, User: m.User})

	return b.String()
}

func (m *Model) buildDisplay() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	var b strings.Builder
	w := b.WriteString

	u := m.User
	if m.err != nil {
		w("no results")
	} else if m.User.Login != "" {
		w(m.buildUserDisplay())

		w(utils.GetNewLines(2))

		w(fmt.Sprintf("%v contributions", u.ActivityGraph.ContributionsCount))

		w(utils.GetNewLines(1))

		w(lipgloss.NewStyle().
			Align(lipgloss.Left).
			Render(graph.BuildGraphDisplay(u.ActivityGraph.Weeks)))

		w(utils.GetNewLines(2))

		w(lipgloss.NewStyle().
			Align(lipgloss.Center).Render(repo.BuildPinnedRepoDisplay(u.PinnedRepos, m.ctx)))

	}

	m.display = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(physicalWidth).Render(b.String())

}

func (m Model) View() string {
	return lipgloss.NewStyle().Height(m.ctx.Layout.ContentHeight).Render(m.display)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}
