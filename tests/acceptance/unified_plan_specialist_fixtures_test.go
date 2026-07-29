// Acceptance: specs/unified-plan-specialist.feature
//
// Shared binary + fixture helpers (Slice 7). Fixtures write workflow/evidence
// JSON directly as strings (matching adversarial_validate_verifier_report_test.go)
// instead of round-tripping through `start`/`evidence init`. No network remote.
package acceptance_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var upsBinOnce sync.Once
var upsBin string

// upsBuildBin compiles the centinela binary ONCE for the whole file set.
func upsBuildBin(t *testing.T) string {
	t.Helper()
	upsBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-ups-bin")
		if err != nil {
			t.Fatal(err)
		}
		upsBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", upsBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return upsBin
}

// upsExistingRepo creates a temp dir configured as an "existing" project so
// `centinela start` skips the roadmap dependency machinery entirely and uses
// workflow.DefaultStepOrder.
func upsExistingRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// upsWorkflowJSON builds a minimal strict-mode plan-step workflow state,
// pinned (or not, when planContract=="") to the given PlanContract.
func upsWorkflowJSON(feature, planContract string) string {
	return fmt.Sprintf(`{"feature":%q,"currentStep":"plan",`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"orchestrationMode":"strict-subagents-v1","planContract":%q,`+
		`"validateContract":"adversarial-v1"}`, feature, planContract)
}

// upsWriteWorkflow writes upsWorkflowJSON for feature at dir.
func upsWriteWorkflow(t *testing.T, dir, feature, planContract string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), upsWorkflowJSON(feature, planContract))
}

// upsFeatureDocs writes the plan-step artifacts validatePlan and
// RequiredPlanInputs need: a feature brief, the plan file, and a Gherkin spec.
func upsFeatureDocs(t *testing.T, dir, feature string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "docs", "features", feature+".md"), "# "+feature+"\n")
	mustWrite(t, filepath.Join(dir, "docs", "plans", feature+".md"), "# Plan\n")
	mustWrite(t, filepath.Join(dir, "specs", feature+".feature"), "Feature: x\n  Scenario: y\n")
}

// jsonArray renders a JSON string array literal.
func jsonArray(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

// upsWriteEvidence writes a schema-valid evidence JSON + companion .md for
// role at the plan step: inputs snapshot the two required plan-input paths,
// outputs point at the plan file (satisfies the plan/spec artifact rule).
func upsWriteEvidence(t *testing.T, dir, feature, role, handoffTo string, edgeCases []string) {
	t.Helper()
	body := fmt.Sprintf(`{"feature":%q,"step":"plan","role":%q,"status":"done",`+
		`"generatedAt":"2026-01-01T00:00:00Z",`+
		`"inputs":["docs/features/%s.md","docs/plans/%s.md"],`+
		`"outputs":["docs/plans/%s.md"],"edgeCases":%s,"handoffTo":%q}`,
		feature, role, feature, feature, feature, jsonArray(edgeCases), handoffTo)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-"+role+".json"), body)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-"+role+".md"), "# "+role+" report\n")
}
