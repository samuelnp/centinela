// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const upsPlannerPromptPath = "docs/architecture/planner-prompt.md"

func upsPlannerPromptLines(t *testing.T) []string {
	t.Helper()
	body := readFile(t, filepath.Join(repoRoot(t), upsPlannerPromptPath))
	return strings.Split(body, "\n")
}

func upsFindLine(lines []string, substr string) int {
	for i, l := range lines {
		if strings.Contains(l, substr) {
			return i
		}
	}
	return -1
}

// Scenario: The planner prompt doc contains both lenses in strategy-then-spec order
func TestUPS_PlannerPromptBothLensesOrdered(t *testing.T) {
	lines := upsPlannerPromptLines(t)
	strategyIdx := upsFindLine(lines, "Lens 1: strategy")
	specIdx := upsFindLine(lines, "Lens 2: spec")
	if strategyIdx < 0 || specIdx < 0 {
		t.Fatalf("must contain both lens headings, strategy=%d spec=%d", strategyIdx, specIdx)
	}
	if strategyIdx >= specIdx {
		t.Fatalf("strategy heading (line %d) must precede spec heading (line %d)", strategyIdx, specIdx)
	}
	body := strings.Join(lines, "\n")
	for _, want := range []string{"## Purpose", "## Prompt Template", "## Required Artifact"} {
		if strings.Count(body, want) != 1 {
			t.Fatalf("must contain exactly one %q heading, got %d", want, strings.Count(body, want))
		}
	}
	if strings.Count(body, "Authoring rules (REQUIRED):") != 1 {
		t.Fatalf("must contain exactly one CLI-authoring-rules block, got %d",
			strings.Count(body, "Authoring rules (REQUIRED):"))
	}
	for _, want := range []string{"#### Deferred Findings", "senior-engineer"} {
		if !strings.Contains(body, want) {
			t.Fatalf("must contain %q: missing from prompt", want)
		}
	}
}

// Scenario: The planner prompt doc respects the line budget
func TestUPS_PlannerPromptLineBudget(t *testing.T) {
	lines := upsPlannerPromptLines(t)
	if len(lines) > 130 {
		t.Fatalf("planner-prompt.md must be at most 130 lines, got %d", len(lines))
	}
}

// Scenario: The legacy prompt docs no longer exist
func TestUPS_LegacyPromptDocsAbsent(t *testing.T) {
	root := repoRoot(t)
	for _, gone := range []string{
		"docs/architecture/big-thinker-prompt.md",
		"docs/architecture/feature-specialist-prompt.md",
	} {
		if _, err := os.Stat(filepath.Join(root, gone)); !os.IsNotExist(err) {
			t.Fatalf("%s must no longer exist", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(root, upsPlannerPromptPath)); err != nil {
		t.Fatalf("planner-prompt.md must exist: %v", err)
	}
}
