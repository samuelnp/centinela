package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/centinela-insights.feature

const (
	advEvt = `{"type":"step-advanced"}`
	rejEvt = `{"type":"complete-rejected"}`
)

func stepsBody(out string) string { return sectionBody(out, "Steps-to-Green") }

func repeat(line string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = line
	}
	return out
}

// Scenario: Mean steps-to-green is computed correctly from known counts
func TestInsightsStepsKnownMean(t *testing.T) {
	log := append(repeat(advEvt, 4), repeat(rejEvt, 2)...)
	out, code := runInsights(t, insightsRepo(t, log))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(stepsBody(out), "1.50") {
		t.Fatalf("expected 1.50: %q", out)
	}
}

// Scenario: Zero step-advanced events renders steps-to-green as n/a without panic
func TestInsightsStepsZeroAdvances(t *testing.T) {
	out, code := runInsights(t, insightsRepo(t, repeat(rejEvt, 3)))
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if !strings.Contains(stepsBody(out), "n/a") {
		t.Fatalf("expected n/a: %q", out)
	}
}

// Scenario: Single step-advanced with no rejections yields mean of 1.00
func TestInsightsStepsSingleAdvance(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{advEvt}))
	if !strings.Contains(stepsBody(out), "1.00") {
		t.Fatalf("expected 1.00: %q", out)
	}
}

// Scenario: Single step-advanced with one rejection yields mean of 2.00
func TestInsightsStepsOneRejection(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, []string{advEvt, rejEvt}))
	if !strings.Contains(stepsBody(out), "2.00") {
		t.Fatalf("expected 2.00: %q", out)
	}
}

// Scenario: Log with only step-advanced events and no complete-rejected reports mean of 1.00
func TestInsightsStepsAllAdvances(t *testing.T) {
	out, _ := runInsights(t, insightsRepo(t, repeat(advEvt, 5)))
	if !strings.Contains(stepsBody(out), "1.00") {
		t.Fatalf("expected 1.00: %q", out)
	}
}
