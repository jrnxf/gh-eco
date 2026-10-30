package user

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/ui/components/graph"
	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/ui/models"
)

func Test_BuildDisplayIsIdempotent(t *testing.T) {
	ctx := &context.ProgramContext{}
	m := NewModel()
	m.UpdateProgramContext(ctx)
	m.User = models.User{
		Login:       "octocat",
		PinnedRepos: []models.Repo{{Id: "r1", Name: "a"}, {Id: "r2", Name: "b"}},
	}
	// BuildGraphDisplay panics on an empty Weeks slice; one empty week keeps
	// buildDisplay from crashing without affecting the widget count assertion.
	m.User.ActivityGraph.Weeks = []models.WeeklyContribution{{}}

	m.buildDisplay()
	first := len(ctx.FocusableWidgets)
	// expect NoFocus + UserDisplay + 2 pinned repos
	if first != 4 {
		t.Fatalf("after first build: got %d widgets, want 4", first)
	}

	m.buildDisplay() // simulates a resize / star-response rebuild
	if got := len(ctx.FocusableWidgets); got != first {
		t.Errorf("after second build: got %d widgets, want %d (rebuild must not duplicate)", got, first)
	}
}

func Test_FitWeeksKeepsGraphInsideWidth(t *testing.T) {
	weeks := make([]models.WeeklyContribution, 53)
	for i := range weeks {
		weeks[i].ContributionDays = []struct{ ContributionLevel string }{{"NONE"}}
	}
	weeks[52].ContributionDays[0].ContributionLevel = "FOURTH_QUARTILE"

	for _, width := range []int{0, 80, 105, 106, 160} {
		got := fitWeeks(weeks, width)
		if width == 0 || width >= 106 {
			if len(got) != len(weeks) {
				t.Errorf("width %d: got %d weeks, want all %d", width, len(got), len(weeks))
			}
			continue
		}
		if len(got) == 0 || got[len(got)-1].ContributionDays[0].ContributionLevel != "FOURTH_QUARTILE" {
			t.Errorf("width %d: most recent week was dropped", width)
		}
		for _, line := range strings.Split(strings.TrimRight(graph.BuildGraphDisplay(got), "\n"), "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Errorf("width %d: graph row is %d cols wide", width, w)
			}
		}
	}
}
