// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// tdReadInputs reads the inputs field of a written evidence JSON.
func tdReadInputs(t *testing.T, dir, feature, role string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".workflow", feature+"-"+role+".json"))
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var doc struct {
		Inputs []string `json:"inputs"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal evidence: %v", err)
	}
	return doc.Inputs
}

// Scenario: evidence init pre-fills exactly the required inputs
func TestTD_EvidenceInitPrefillsExactRequiredInputs(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteWorkflow(t, dir, "token-diet", "plan", "planner-v1", false)
	tdFeatureDocs(t, dir, "token-diet")

	tdRun(t, bin, dir, "evidence", "init", "token-diet", "planner")
	got := tdReadInputs(t, dir, "token-diet", "planner")
	want := orchestration.RequiredPlanInputs("token-diet")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("prefill = %v, want exactly %v", got, want)
	}

	// Complete the remaining required fields via the CLI (not by hand-editing
	// JSON) and confirm the snapshot rule specifically is satisfied.
	tdRun(t, bin, dir, "evidence", "append", "token-diet", "planner", "outputs", "docs/plans/token-diet.md")
	tdRun(t, bin, dir, "evidence", "append", "token-diet", "planner", "edgeCases", "prefill covers both required paths")
	out, code := runCent(t, bin, dir, "evidence", "validate", "token-diet")
	if code != 0 {
		t.Fatalf("init->append->validate should exit 0: %s", out)
	}
	mustNotContain(t, out, "missing feature-doc snapshot inputs")
}

// Scenario: A retired legacy plan role gets the same shrunken set
func TestTD_RetiredLegacyPlanRoleGetsSameShrunkenSet(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	// Unpinned (planContract=="") in-flight workflow: legacy roles are still
	// allowed, unlike a workflow pinned to planner-v1.
	tdWriteWorkflow(t, dir, "already-in-flight", "plan", "", false)
	tdFeatureDocs(t, dir, "already-in-flight")

	for _, role := range []string{"big-thinker", "feature-specialist"} {
		tdRun(t, bin, dir, "evidence", "init", "already-in-flight", role)
		got := tdReadInputs(t, dir, "already-in-flight", role)
		want := orchestration.RequiredPlanInputs("already-in-flight")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("role %s: prefill = %v, want exactly %v", role, got, want)
		}
	}
}
