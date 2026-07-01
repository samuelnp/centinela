// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// rlChdirWorkflow isolates into a temp project with an empty .workflow/.
func rlChdirWorkflow(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
}

// Scenario: Re-gating — complete at re-opened step is blocked until evidence is regenerated
func TestRL_ReGatingBlocksWithoutEvidence(t *testing.T) {
	rlChdirWorkflow(t)
	// After a rewind, the re-opened tests step's qa-senior evidence is gone, so
	// the gate the next `complete` runs reports it missing and refuses to advance.
	err := orchestration.ValidateStep("my-feature", "tests", nil)
	if err == nil || !strings.Contains(err.Error(), "qa-senior") {
		t.Fatalf("missing qa-senior evidence must block: %v", err)
	}
}

// Scenario: Re-gating — complete advances once evidence is regenerated
func TestRL_ReGatingAdvancesAfterRegenerated(t *testing.T) {
	rlChdirWorkflow(t)
	if err := os.MkdirAll("tests/unit", 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(p, body string) {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tests/unit/x_test.go", "package x")
	write(".workflow/my-feature-edge-cases.md", "edge")
	write(".workflow/my-feature-qa-senior.md", "# qa")
	write(".workflow/my-feature-qa-senior.json", `{"feature":"my-feature",`+
		`"step":"tests","role":"qa-senior","status":"done",`+
		`"generatedAt":"2026-06-30T00:00:00Z","inputs":["docs/plans/my-feature.md"],`+
		`"outputs":["tests/unit/x_test.go",".workflow/my-feature-edge-cases.md"],`+
		`"edgeCases":["empty input"],"handoffTo":"validation-specialist"}`)
	if err := orchestration.ValidateStep("my-feature", "tests", nil); err != nil {
		t.Fatalf("regenerated evidence must pass the gate: %v", err)
	}
}
