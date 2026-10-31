package graph

import (
	"strings"
	"testing"

	"github.com/jrnxf/gh-eco/models"
)

func Test_BuildGraphDisplay_Empty(t *testing.T) {
	if output := BuildGraphDisplay(nil); output != "" {
		t.Errorf("expected empty string for nil input, got %q", output)
	}
}

func Test_BuildGraphDisplay_ZeroDays(t *testing.T) {
	if output := BuildGraphDisplay([]models.WeeklyContribution{{}}); output != "" {
		t.Errorf("expected empty string for zero-day weeks, got %q", output)
	}
}

func Test_BuildGraphDisplay_Ragged(t *testing.T) {
	weeks := []models.WeeklyContribution{
		{ContributionDays: []struct{ ContributionLevel string }{{"NONE"}, {"FIRST_QUARTILE"}}},
		{ContributionDays: []struct{ ContributionLevel string }{{"NONE"}, {"NONE"}, {"FOURTH_QUARTILE"}}},
	}

	output := BuildGraphDisplay(weeks)
	if output == "" {
		t.Errorf("expected non-empty output for ragged weeks, got empty string")
	}
}

func Test_BuildGraphDisplay_Shape(t *testing.T) {
	weeks := []models.WeeklyContribution{
		{ContributionDays: []struct{ ContributionLevel string }{{"NONE"}, {"FIRST_QUARTILE"}}},
		{ContributionDays: []struct{ ContributionLevel string }{{"SECOND_QUARTILE"}, {"FOURTH_QUARTILE"}}},
	}

	output := BuildGraphDisplay(weeks)
	if count := strings.Count(output, "\n"); count != 2 {
		t.Errorf("expected 2 newlines for a 2x2 input, got %d (output %q)", count, output)
	}
}
