package user

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coloradocolby/ghx/api"
	"github.com/coloradocolby/ghx/utils"
	"golang.org/x/term"
)

var (
	// colors
	subtleC    = lipgloss.Color("#717c93")
	highlightC = lipgloss.Color("#81A1C1")
)

var (
	subtle = lipgloss.NewStyle().Foreground(subtleC).Render
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
		m.User = msg.User
		m.buildUi()
	}
	return m, cmd
}

func (m *Model) buildUi() {
	log.Println("s")
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	var b strings.Builder
	w := b.WriteString

	u := m.User

	w(bold(u.Name) + "\n\n")

	w(u.Bio + "\n\n")

	w(fmt.Sprintf("%v %v · %v %v\n\n", u.Followers.TotalCount, "followers", u.Following.TotalCount, "following"))

	w(fmt.Sprintf("%v  |  %v  |  @%v\n\n\n", u.Location, u.WebsiteUrl, u.TwitterUsername))

	w(fmt.Sprintf("%v contributions\n", m.User.ContributionsCollection.ContributionCalendar.TotalContributions))
	w(utils.ConvertWeeklyContributionsToGraph(u.ContributionsCollection.ContributionCalendar.Weeks))

	m.display = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(physicalWidth).Render(b.String())
	log.Println("e")

}

func (m Model) View() string {
	return m.display
}
