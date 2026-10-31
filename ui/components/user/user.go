package user

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/graph"
	"github.com/jrnxf/gh-eco/ui/components/repo"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
)

// rightSlack keeps every line this many cells short of the screen edge.
// Some emoji draw wider than lipgloss measures them; a line that reached the
// last column would then wrap and shift the whole frame down.
const rightSlack = 2

// maxErrorWidth wraps long error messages to a readable width.
const maxErrorWidth = 60

// wheelLines is how far one mouse wheel step scrolls.
const wheelLines = 3

type Model struct {
	lines []string
	// login is the latest search.
	login string
	// offset is the first visible line.
	offset int
	// focus is the focus the scroll position was last fitted to.
	focus int
	err   error
	state *state.State
}

func NewModel(s *state.State) Model {
	return Model{state: s, focus: -1}
}

// Update rebuilds the display after every message, since the root model may
// have changed the user, the focus or the layout.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.GetUserResponse:
		m.login, m.err = msg.Login, msg.Err
		m.offset = 0
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.offset -= wheelLines
			case tea.MouseButtonWheelDown:
				m.offset += wheelLines
			}
		}
	}
	m.build()
	return m, nil
}

func (m *Model) build() {
	width := max(0, m.state.Layout.ContentWidth-rightSlack)
	u := m.state.User

	var blocks []string
	switch {
	case errors.Is(m.err, github.ErrNotFound):
		blocks = append(blocks, ansi.Truncate(fmt.Sprintf("no user or organization named %q", m.login), width, "..."))
	case m.err != nil:
		blocks = append(blocks, lipgloss.NewStyle().
			Width(min(width, maxErrorWidth)).
			Render(fmt.Sprintf("couldn't load %q: %v", m.login, m.err)))
	case u.Login != "":
		blocks = append(blocks, m.header(width)...)
		blocks = append(blocks, "")
		if !u.IsOrganization {
			blocks = append(blocks,
				fmt.Sprintf("%v contributions", u.ActivityGraph.ContributionsCount),
				graph.BuildGraphDisplay(fitWeeks(u.ActivityGraph.Weeks, width)),
				"",
			)
		}
	}

	var lines []string
	for _, b := range blocks {
		lines = append(lines, centerBlock(b, width)...)
	}

	focusTop, focusBottom := 0, 0
	if u.Login != "" && m.err == nil {
		focused := -1
		if t := m.state.Focused(); t.Kind == state.RepoTarget {
			focused = t.Repo
		}
		grid, top, bottom := repo.Grid(u.PinnedRepos, focused, width)
		if top >= 0 {
			focusTop, focusBottom = len(lines)+top, len(lines)+bottom
		}
		lines = append(lines, centerBlock(grid, width)...)
	}
	m.lines = lines

	if m.focus != m.state.Focus {
		// scroll just enough to show the newly focused element
		m.focus = m.state.Focus
		height := m.state.Layout.ContentHeight
		if focusTop < m.offset {
			m.offset = focusTop
		} else if focusBottom >= m.offset+height {
			m.offset = focusBottom - height + 1
		}
	}
	m.offset = max(0, min(m.offset, len(m.lines)-m.state.Layout.ContentHeight))
}

// header is the name, bio, follower counts and links, one block each.
func (m Model) header(width int) []string {
	u := m.state.User

	name := u.Name
	if name == "" {
		// user hasn't specified their name on github
		name = u.Login
	}
	nameStyle := styles.Bold
	if m.state.Focused().Kind == state.UserTarget {
		nameStyle = styles.FocusedBold
	}
	blocks := []string{nameStyle.Render(ansi.Truncate(name, width, "...")), ""}

	if u.Bio != "" {
		bio := lipgloss.NewStyle().
			Faint(true).
			Align(lipgloss.Center).
			Width(min(width, lipgloss.Width(u.Bio))).
			Render(u.Bio)
		blocks = append(blocks, bio, "")
	}

	if !u.IsOrganization {
		follow := fmt.Sprintf("%v followers", u.FollowersCount)
		if u.ViewerIsFollowing {
			follow += lipgloss.NewStyle().Italic(true).Render(" (you follow)")
		}
		follow += fmt.Sprintf(" / %v following", u.FollowingCount)
		if u.IsFollowingViewer {
			follow += lipgloss.NewStyle().Italic(true).Render(" (follows you)")
		}
		blocks = append(blocks, ansi.Truncate(follow, width, "..."))
	}

	var links []string
	if u.Location != "" {
		links = append(links, u.Location)
	}
	if u.WebsiteUrl != "" {
		links = append(links, u.WebsiteUrl)
	}
	if u.TwitterUsername != "" {
		links = append(links, "@"+u.TwitterUsername)
	}
	if len(links) > 0 {
		if !u.IsOrganization {
			blocks = append(blocks, "")
		}
		blocks = append(blocks, ansi.Truncate(strings.Join(links, "  /  "), width, "..."))
	}
	if blocks[len(blocks)-1] == "" {
		blocks = blocks[:len(blocks)-1]
	}
	return blocks
}

// centerBlock indents every line of block by the same amount so the block
// sits centered in width. Unlike lipgloss alignment it adds no trailing
// padding, so no line runs to the screen edge.
func centerBlock(block string, width int) []string {
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	pad := strings.Repeat(" ", max(0, (width-lipgloss.Width(block))/2))
	for i, l := range lines {
		if l != "" {
			lines[i] = pad + l
		}
	}
	return lines
}

// fitWeeks keeps the most recent weeks that fit in width, since each week
// renders as a 2-col graph column and wider rows would wrap.
func fitWeeks(weeks []models.WeeklyContribution, width int) []models.WeeklyContribution {
	if maxWeeks := width / 2; width > 0 && len(weeks) > maxWeeks {
		return weeks[len(weeks)-maxWeeks:]
	}
	return weeks
}

// View shows the visible slice of the display, exactly ContentHeight lines.
func (m Model) View() string {
	height := m.state.Layout.ContentHeight
	visible := make([]string, height)
	if m.offset < len(m.lines) {
		copy(visible, m.lines[m.offset:])
	}
	return strings.Join(visible, "\n")
}
