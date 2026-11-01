// Package people shows a user's followers or the people they follow.
package people

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/state"
	"github.com/jrnxf/gh-eco/ui/styles"
	"github.com/jrnxf/gh-eco/ui/text"
)

// maxRowWidth caps the list column, so on a wide screen each name stays
// next to its login.
const maxRowWidth = 64

// headerLines is the title and the blank line under it.
const headerLines = 2

var selectedLogin = lipgloss.NewStyle().Bold(true).Foreground(styles.Accent)

type Model struct {
	state *state.State
}

func NewModel(s *state.State) Model {
	return Model{state: s}
}

// Update scrolls the list just enough to keep the selection on screen,
// since the root model may have moved it or resized the screen.
func (m Model) Update(tea.Msg) (Model, tea.Cmd) {
	p := &m.state.People
	rows := m.rows()
	if p.Selected < p.Offset {
		p.Offset = p.Selected
	} else if p.Selected >= p.Offset+rows {
		p.Offset = p.Selected - rows + 1
	}
	p.Offset = max(0, min(p.Offset, len(p.List)-rows))
	return m, nil
}

// rows is how many entries fit under the title.
func (m Model) rows() int {
	return max(1, m.state.Layout.ContentHeight-headerLines)
}

// View shows the title and the visible entries, exactly ContentHeight
// lines. A non-empty spinner goes where the next entries will appear, once
// the list is scrolled to its end.
func (m Model) View(spinner string) string {
	width := max(0, m.state.Layout.ContentWidth-state.RightSlack)
	p := m.state.People

	lines := []string{center(title(p), width), ""}
	switch {
	case len(p.List) > 0:
		column := min(width, maxRowWidth)
		pad := strings.Repeat(" ", (width-column)/2)
		start, end := p.Offset, min(len(p.List), p.Offset+m.rows())
		atEnd := spinner != "" && end == len(p.List)
		if atEnd && end-start == m.rows() {
			// the spinner takes an end row the selection is not on
			if p.Selected == end-1 {
				start++
			} else {
				end--
			}
		}
		for i := start; i < end; i++ {
			lines = append(lines, pad+row(p.List[i], i == p.Selected, column))
		}
		if atEnd {
			lines = append(lines, center(spinner, width))
		}
	case spinner != "":
		lines = append(lines, center(spinner, width))
	case p.Failed:
		lines = append(lines, center("couldn't load "+Noun(p.Kind), width))
	case p.Loaded:
		empty := "no followers yet"
		if p.Kind == models.Following {
			empty = "not following anyone yet"
		}
		lines = append(lines, center(empty, width))
	}

	visible := make([]string, m.state.Layout.ContentHeight)
	copy(visible, lines)
	return strings.Join(visible, "\n")
}

// title names whose list it is and how long it is, worded like the
// profile's counts.
func title(p state.People) string {
	return text.Count(p.Total) + " " + Noun(p.Kind) + styles.Muted().Render(" · "+p.Login)
}

// Noun names a list as the profile's counts do.
func Noun(kind models.PeopleKind) string {
	if kind == models.Following {
		return "following"
	}
	return "followers"
}

// row is one entry, its login then its name, cut to width.
func row(person models.Person, selected bool, width int) string {
	marker, login := "  ", person.Login
	if selected {
		marker, login = selectedLogin.Render("› "), selectedLogin.Render(person.Login)
	}
	line := marker + login
	if person.Name != "" && person.Name != person.Login {
		line += "  " + styles.Muted().Render(person.Name)
	}
	return text.Ellipsize(line, width)
}

// center indents s to sit centered in width, with no trailing padding.
func center(s string, width int) string {
	s = text.Ellipsize(s, width)
	return strings.Repeat(" ", max(0, (width-lipgloss.Width(s))/2)) + s
}
