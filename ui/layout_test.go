package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
		u.PinnedRepos = append(u.PinnedRepos, r)
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

func checkFrame(t *testing.T, view string, width, height, slack int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("frame is %d lines, screen is %d", len(lines), height)
	}
	for i, line := range lines[:len(lines)-1] { // the help bar may use every column
		if w := lipgloss.Width(line); w > width-slack {
			t.Errorf("line %d is %d cells, limit %d: %q", i, w, width-slack, line)
		}
	}
}

func Test_FrameFitsScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {160, 50}, {50, 15}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			client := &fakeClient{
				users:   map[string]models.User{"octocat": busyUser()},
				readmes: map[string]string{"octocat/octocat": "# hi\n\n🤘🏼 " + strings.Repeat("word ", 300)},
			}
			m, _ := send(New(client), tea.WindowSizeMsg{Width: width, Height: height})
			m, res := lookup(t, m, "octocat")
			m, _ = send(m, res)

			// user view: no line may come within two cells of the edge,
			// where an emoji drawn wider than measured would wrap
			checkFrame(t, m.View(), width, height, 2)
			if !strings.HasPrefix(strings.Split(m.View(), "\n")[1], "> octocat") {
				t.Errorf("search bar scrolled off the top")
			}

			// walking focus to the last pin scrolls it into view
			for range len(busyUser().PinnedRepos) + 1 {
				m, _ = send(m, keyRunes("j"))
			}
			checkFrame(t, m.View(), width, height, 2)
			if !strings.Contains(m.View(), "repo-5") {
				t.Errorf("focused repo-5 is not on screen")
			}

			// readme view
			m, _ = send(m, keyRunes("k"))
			for m.state.Focused().Kind != state.UserTarget {
				m, _ = send(m, keyRunes("k"))
			}
			m, msgs := send(m, keyRunes("r"))
			m, _ = send(m, find[commands.GetReadmeResponse](t, msgs))
			if m.state.CurrentView != state.ReadmeView {
				t.Fatalf("readme did not open")
			}
			checkFrame(t, m.View(), width, height, 0)
		})
	}
}

func Test_ResizeKeepsFocusedCardVisible(t *testing.T) {
	client := &fakeClient{users: map[string]models.User{"octocat": busyUser()}}
	m, _ := send(New(client), tea.WindowSizeMsg{Width: 160, Height: 50})
	m, res := lookup(t, m, "octocat")
	m, _ = send(m, res)
	for range len(busyUser().PinnedRepos) + 1 {
		m, _ = send(m, keyRunes("j"))
	}
	if r := m.state.FocusedRepo(); r == nil || r.Name != "repo-5" {
		t.Fatalf("focused repo = %+v, want repo-5", r)
	}

	for _, size := range [][2]int{{120, 30}, {50, 18}, {36, 14}, {160, 50}} {
		width, height := size[0], size[1]
		m, _ = send(m, tea.WindowSizeMsg{Width: width, Height: height})
		checkFrame(t, m.View(), width, height, 2)
		if !strings.Contains(m.View(), "repo-5") {
			t.Errorf("%dx%d: focused repo-5 is not on screen", width, height)
		}
	}
}
