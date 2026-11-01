package user

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/graph"
	"github.com/jrnxf/gh-eco/ui/components/repo"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
	"github.com/jrnxf/gh-eco/ui/text"
)

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
	// focus and size are the focus and content size the scroll position
	// was last fitted to.
	focus int
	size  [2]int
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
	width := max(0, m.state.Layout.ContentWidth-state.RightSlack)
	u := m.state.User

	var blocks []string
	switch {
	case errors.Is(m.err, github.ErrNotFound):
		blocks = append(blocks, text.Ellipsize(fmt.Sprintf("no user or organization named %q", m.login), width))
	case m.err != nil:
		blocks = append(blocks, lipgloss.NewStyle().
			Width(min(width, maxErrorWidth)).
			Render(fmt.Sprintf("couldn't load %q: %v", m.login, m.err)))
	case u.Login != "":
		blocks = append(blocks, m.header(width)...)
		blocks = append(blocks, "")
		if !u.IsOrganization {
			blocks = append(blocks,
				text.Ellipsize(countOf(u.ActivityGraph.ContributionsCount)+" contributions in the last year", width),
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
	gridTop, rowHeight := 0, 0
	if u.Login != "" && m.err == nil {
		focused := -1
		if t := m.state.Focused(); t.Kind == state.RepoTarget {
			focused = t.Repo
		}
		grid, height, top, bottom := repo.Grid(u.PinnedRepos, focused, width)
		gridTop, rowHeight = len(lines), height
		if top >= 0 {
			focusTop, focusBottom = len(lines)+top, len(lines)+bottom
		}
		lines = append(lines, centerBlock(grid, width)...)
	}
	m.lines = lines

	size := [2]int{m.state.Layout.ContentWidth, m.state.Layout.ContentHeight}
	if m.focus != m.state.Focus || m.size != size {
		// scroll just enough to show the focused element
		m.focus, m.size = m.state.Focus, size
		height := m.state.Layout.ContentHeight
		if focusTop < m.offset {
			m.offset = focusTop
		} else if focusBottom >= m.offset+height {
			m.offset = min(snapToRow(focusBottom-height+1, gridTop, rowHeight), focusTop)
		}
	}
	// the last screenful may end in blank lines so that it, too, starts on
	// a whole row of cards
	maxOffset := snapToRow(len(m.lines)-m.state.Layout.ContentHeight, gridTop, rowHeight)
	m.offset = max(0, min(m.offset, maxOffset))
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
	blocks := []string{nameStyle.Render(text.Ellipsize(name, width))}

	if u.Bio != "" {
		bio := lipgloss.NewStyle().
			Faint(true).
			Align(lipgloss.Center).
			Width(min(width, lipgloss.Width(u.Bio))).
			Render(u.Bio)
		blocks = append(blocks, bio)
	}

	italic := lipgloss.NewStyle().Italic(true)
	var stats []string
	if !u.IsOrganization {
		followers := countOf(u.FollowersCount) + " followers"
		if u.ViewerIsFollowing {
			followers += italic.Render(" (you follow)")
		}
		following := countOf(u.FollowingCount) + " following"
		if u.IsFollowingViewer {
			following += italic.Render(" (follows you)")
		}
		stats = append(stats, followers, following)
	}
	if u.Location != "" {
		stats = append(stats, u.Location)
	}
	if u.WebsiteUrl != "" {
		stats = append(stats, u.WebsiteUrl)
	}
	if u.TwitterUsername != "" {
		stats = append(stats, "@"+u.TwitterUsername)
	}
	if len(stats) > 0 {
		blocks = append(blocks, text.Ellipsize(strings.Join(stats, " · "), width))
	}
	return blocks
}

// countOf formats n with thousands separators, as GitHub does.
func countOf(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + countOf(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// snapToRow moves an offset that falls inside the card grid down to the
// next row boundary, so the top line is never a stray half card.
func snapToRow(offset, gridTop, rowHeight int) int {
	if rowHeight == 0 || offset <= gridTop {
		return offset
	}
	rows := (offset - gridTop + rowHeight - 1) / rowHeight
	return gridTop + rows*rowHeight
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
