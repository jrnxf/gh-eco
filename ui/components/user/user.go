package user

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/graph"
	"github.com/jrnxf/gh-eco/ui/components/repo"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
	"github.com/jrnxf/gh-eco/utils"
	"golang.org/x/term"
)

type Model struct {
	display string
	err     error
	state   *state.State
}

func NewModel(s *state.State) Model {
	return Model{state: s}
}

// Update rebuilds the display after every message, since the root model may
// have changed the user, the focus or the layout.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(commands.GetUserResponse); ok {
		m.err = msg.Err
	}
	m.buildDisplay()
	return m, nil
}

func (m Model) buildUserDisplay() string {
	u := m.state.User
	focused := m.state.Focused().Kind == state.UserTarget

	var b strings.Builder
	w := b.WriteString

	name := u.Name
	if name == "" {
		// user hasn't specified their name on github
		name = u.Login
	}
	if focused {
		w(styles.FocusedBold.Render(name))
	} else {
		w(styles.Bold.Render(name))
	}
	w(utils.GetNewLines(2))

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

	return b.String()
}

func (m *Model) buildDisplay() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	var b strings.Builder
	w := b.WriteString

	u := m.state.User
	if m.err != nil {
		w("no results")
	} else if u.Login != "" {
		w(m.buildUserDisplay())

		w(utils.GetNewLines(2))

		w(fmt.Sprintf("%v contributions", u.ActivityGraph.ContributionsCount))

		w(utils.GetNewLines(1))

		w(lipgloss.NewStyle().
			Align(lipgloss.Left).
			Render(graph.BuildGraphDisplay(fitWeeks(u.ActivityGraph.Weeks, physicalWidth))))

		w(utils.GetNewLines(2))

		focusedRepo := -1
		if t := m.state.Focused(); t.Kind == state.RepoTarget {
			focusedRepo = t.Repo
		}
		w(lipgloss.NewStyle().
			Align(lipgloss.Center).Render(repo.BuildPinnedRepoDisplay(u.PinnedRepos, focusedRepo)))
	}

	m.display = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(physicalWidth).Render(b.String())
}

// fitWeeks keeps the most recent weeks that fit in width, since each week
// renders as a 2-col graph column and wider rows would wrap.
func fitWeeks(weeks []models.WeeklyContribution, width int) []models.WeeklyContribution {
	if maxWeeks := width / 2; width > 0 && len(weeks) > maxWeeks {
		return weeks[len(weeks)-maxWeeks:]
	}
	return weeks
}

func (m Model) View() string {
	return lipgloss.NewStyle().Height(m.state.Layout.ContentHeight).Render(m.display)
}
