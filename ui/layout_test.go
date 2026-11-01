package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/state"
)

// busyUser has everything that once broke the layout: emoji whose drawn
// width varies by terminal, a repo without a language, a long bio, six pins
// and a full year of contributions.
func busyUser() models.User {
	u := models.User{
		Login:          "octocat",
		Name:           "Mona Octocat",
		Bio:            strings.Repeat("builds terminal tools and writes about them ", 4),
		Location:       "Lisbon",
		WebsiteUrl:     "https://example.com",
		FollowersCount: 329965,
	}
	descriptions := []string{
		"✨ sleek typing tui with visualized results and historical logging",
		"🤘🏼 all things une",
		"📚 flow state reading in the terminal",
		"",
		"A HUD from the shell for macOS: one command shows a brief overlay",
		"漢字の説明文がここに入ります",
	}
	for i, d := range descriptions {
		r := models.Repo{
			Id:          fmt.Sprintf("R_%d", i),
			Name:        fmt.Sprintf("repo-%d", i),
			Description: d,
			StarsCount:  i * 100,
			Owner:       struct{ Login string }{"octocat"},
		}
		if i != 3 {
			r.PrimaryLanguage.Name, r.PrimaryLanguage.Color = "Go", "#00ADD8"
		}
		u.Repos = append(u.Repos, r)
	}
	for range 53 {
		var w models.WeeklyContribution
		for range 7 {
			w.ContributionDays = append(w.ContributionDays, struct{ ContributionLevel string }{"SECOND_QUARTILE"})
		}
		u.ActivityGraph.Weeks = append(u.ActivityGraph.Weeks, w)
	}
	return u
}

// checkFrame asserts the frame is exactly the screen height, that content
// rows keep state.RightSlack cells free at the edge (where an emoji drawn
// wider than measured would otherwise wrap), and that the bottom row fits.
func checkFrame(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("%dx%d: frame is %d lines", width, height, len(lines))
	}
	limit := width - state.RightSlack
	for i, line := range lines[:len(lines)-1] {
		if w := lipgloss.Width(line); w > limit {
			t.Fatalf("%dx%d: line %d is %d cells, limit %d: %q", width, height, i, w, limit, line)
		}
	}
	if w := lipgloss.Width(lines[len(lines)-1]); w > width {
		t.Fatalf("%dx%d: bottom row is %d cells: %q", width, height, w, lines[len(lines)-1])
	}
}

func Test_FrameFitsScreen(t *testing.T) {
	client := &fakeClient{
		users:   map[string]models.User{"octocat": busyUser()},
		readmes: map[string]string{"octocat/octocat": "# hi\n\n🤘🏼 " + strings.Repeat("word ", 300)},
		people: map[string][]models.Person{"followers octocat": {
			{Login: "a-rather-long-login-name-indeed", Name: "🤘🏼 " + strings.Repeat("Very Long Name ", 6)},
			{Login: "kanji", Name: "漢字の名前"},
			{Login: "plain"},
		}},
	}
	for width := 30; width <= 200; width++ {
		heights := []int{24}
		if width%10 == 0 {
			heights = append(heights, 15)
		}
		if width == 60 {
			heights = append(heights, 20)
		}
		for _, height := range heights {
			m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: width, Height: height})
			m, res := lookup(t, m, "octocat")
			m, _ = send(m, res)

			checkFrame(t, m.View(), width, height)
			if trail, _ := topRows(m); trail != "octocat" {
				t.Fatalf("%dx%d: hop trail scrolled off the top: %q", width, height, trail)
			}

			// the focused profile name, with its colored underline
			m, _ = send(m, keyRunes("j"))
			checkFrame(t, m.View(), width, height)

			// walking focus to the last pin scrolls it into view
			m = focusRepo(t, m, "R_5")
			checkFrame(t, m.View(), width, height)
			if !strings.Contains(m.View(), "repo-5") {
				t.Fatalf("%dx%d: focused repo-5 is not on screen", width, height)
			}

			// readme view
			for m.state.Focused().Kind != state.UserTarget {
				m, _ = send(m, keyRunes("k"))
			}
			m, msgs := send(m, keyRunes("r"))
			m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
			if m.state.CurrentView != state.ReadmeView {
				t.Fatalf("%dx%d: readme did not open", width, height)
			}
			checkFrame(t, m.View(), width, height)

			// followers list
			m, _ = send(m, keyRunes("r"))
			m = openList(t, m, state.FollowersTarget)
			checkFrame(t, m.View(), width, height)
			if !strings.Contains(m.View(), "plain") {
				t.Fatalf("%dx%d: followers list is cut off:\n%s", width, height, m.View())
			}
		}
	}
}

func Test_ScrollStartsOnACardRow(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": busyUser()}}
	m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: 80, Height: 24})
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	for range len(m.state.Targets()) - 1 {
		m, _ = send(m, keyRunes("j"))
		// the first content row is never the bottom half of a card
		first := strings.TrimSpace(strings.Split(m.View(), "\n")[2])
		if strings.HasPrefix(first, "│") || strings.HasPrefix(first, "╰") {
			t.Fatalf("focus %d: content starts mid-card: %q", m.state.Focus, first)
		}
	}
}

func Test_ResizeKeepsFocusedCardVisible(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": busyUser()}}
	m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: 160, Height: 50})
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	m = focusRepo(t, m, "R_5")
	if r := m.state.FocusedRepo(); r == nil || r.Name != "repo-5" {
		t.Fatalf("focused repo = %+v, want repo-5", r)
	}

	for _, size := range [][2]int{{120, 30}, {50, 18}, {36, 14}, {160, 50}} {
		width, height := size[0], size[1]
		m, _ = send(m, tea.WindowSizeMsg{Width: width, Height: height})
		checkFrame(t, m.View(), width, height)
		if !strings.Contains(m.View(), "repo-5") {
			t.Errorf("%dx%d: focused repo-5 is not on screen", width, height)
		}
	}
}

func Test_FocusNeverMovesTheLayout(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": busyUser()}}
	m, _ := send(New(client, ""), tea.WindowSizeMsg{Width: 160, Height: 50})
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)

	// rowOf finds the screen row holding s
	rowOf := func(view, s string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(ansi.Strip(line), s) {
				return i
			}
		}
		t.Fatalf("%q not on screen", s)
		return -1
	}
	marks := []string{"Mona Octocat", "builds terminal tools", "contributions", "repo-0", "repo-5"}
	before := map[string]int{}
	for _, s := range marks {
		before[s] = rowOf(m.View(), s)
	}
	for range len(m.state.Targets()) - 1 { // the profile, its counts, then every card
		m, _ = send(m, keyRunes("j"))
		for _, s := range marks {
			if got := rowOf(m.View(), s); got != before[s] {
				t.Fatalf("focus %d moved %q from row %d to %d", m.state.Focus, s, before[s], got)
			}
		}
	}
}
