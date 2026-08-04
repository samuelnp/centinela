// Acceptance: specs/hook-context-panel-diet.feature
//
// The size-guard scenarios. These drive the colocated guard
// (internal/ui/panel_budget_test.go) as a subprocess, because the thing under
// test is the guard's own pass/fail behaviour — asserting a copy of its logic
// here would prove nothing about the check that actually runs in CI.
package acceptance_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const pdBudgetTest = "TestPanelBudgetWithinByteBudget"

// Scenario: The size guard passes at the recorded budget
func TestPDSizeGuardPassesAtRecordedBudget(t *testing.T) {
	out, code := pdRunGuard(t, "TestPanelBudget", "")
	if code != 0 {
		t.Fatalf("size guard should pass unpadded, exited %d:\n%s", code, out)
	}
	mustContain(t, out, "PASS: "+pdBudgetTest)
	mustContain(t, out, "PASS: TestPanelBudgetNoTrailingWhitespace")
}

// Scenario: The size guard fails when panel output grows past budget
//
// The guard reads CENTINELA_PANEL_DIET_PAD (test-only; no production code
// touches it) so this scenario can watch it go red for real rather than
// trusting that it would.
func TestPDSizeGuardFailsPastBudget(t *testing.T) {
	const pad = 500
	out, code := pdRunGuard(t, pdBudgetTest, strconv.Itoa(pad))
	if code == 0 {
		t.Fatalf("size guard stayed green with %d bytes of injected padding — it is "+
			"not measuring what it claims:\n%s", pad, out)
	}
	mustContain(t, out, "hook panel budget exceeded")
	// The failure must name both numbers, or the next person has to go
	// re-measure by hand to find out how far past budget they are. Read them
	// back out of the message rather than hardcoding today's values, so a
	// legitimate re-baseline does not break this scenario.
	m := pdBudgetNumbers.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("failure message must state the measured byte length and the budget:\n%s", out)
	}
	measured, budget := pdAtoi(t, m[1]), pdAtoi(t, m[2])
	if measured <= budget {
		t.Fatalf("failure message is self-contradictory: measured %d, budget %d", measured, budget)
	}
	if measured < pad {
		t.Fatalf("measured %d bytes is smaller than the %d injected — the pad knob is a no-op", measured, pad)
	}
}

var pdBudgetNumbers = regexp.MustCompile(`measured (\d+) bytes, budget (\d+)`)

func pdAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("unparseable byte count %q: %v", s, err)
	}
	return n
}

// Scenario: The size guard does not depend on this repository's live workflow count
//
// Delegates to TestPanelBudgetIndependentOfRepoState, which re-measures the
// fixture from a directory holding seven unrelated workflow state files and
// requires a byte-identical result — the "two different machines" property.
func TestPDSizeGuardIgnoresLiveWorkflowCount(t *testing.T) {
	out, code := pdRunGuard(t, "TestPanelBudgetIndependentOfRepoState", "")
	if code != 0 {
		t.Fatalf("machine-independence guard failed, exited %d:\n%s", code, out)
	}
	mustContain(t, out, "PASS: TestPanelBudgetIndependentOfRepoState")

	// The guard must also be free of live-state reads by construction, not just
	// by outcome: no active-workflow lookup, no .workflow scan.
	src := pdReadGuardSource(t)
	for _, forbidden := range []string{"ActiveWorkflows", "workflow.Load", "os.ReadDir"} {
		if strings.Contains(src, forbidden) {
			t.Errorf("size guard reads live state via %q — it must measure the fixed "+
				"in-memory fixture only", forbidden)
		}
	}
}
