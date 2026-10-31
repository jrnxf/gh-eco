package user

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jrnxf/gh-eco/models"
	"github.com/jrnxf/gh-eco/ui/components/graph"
)

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
