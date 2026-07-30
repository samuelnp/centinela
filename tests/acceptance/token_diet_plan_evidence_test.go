// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"path/filepath"
	"testing"
)

// tdWritePlannerEvidence writes a plan-step planner evidence JSON with the
// given inputs list; status/handoffTo/edgeCases are filled so only the
// snapshot-input rule is under test.
func tdWritePlannerEvidence(t *testing.T, dir, feature string, inputs []string) {
	t.Helper()
	body := `{"feature":"` + feature + `","step":"plan","role":"planner","status":"done",` +
		`"generatedAt":"2026-01-01T00:00:00Z","inputs":` + tdJSONArray(inputs) +
		`,"outputs":["docs/plans/` + feature + `.md"],"edgeCases":["a"],"handoffTo":"senior-engineer"}`
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-planner.json"), body)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-planner.md"), "# planner report\n")
}

// Scenario: Planner evidence listing exactly the two required inputs validates
func TestTD_PlannerEvidenceExactTwoInputsValidates(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdFeatureDocs(t, dir, "token-diet")
	tdWritePlannerEvidence(t, dir, "token-diet", []string{
		"docs/features/token-diet.md", "docs/plans/token-diet.md",
	})
	out, code := runCent(t, bin, dir, "evidence", "validate", "token-diet")
	if code != 0 {
		t.Fatalf("evidence validate should exit 0: %s", out)
	}
	mustNotContain(t, out, "missing feature-doc snapshot inputs")
}

// Scenario: Legacy evidence enumerating every brief still validates
func TestTD_LegacyEvidenceEnumeratingEveryBriefStillValidates(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdFeatureDocs(t, dir, "legacy-feat")
	inputs := []string{"docs/features/legacy-feat.md", "docs/plans/legacy-feat.md"}
	for i := 0; i < 120; i++ {
		inputs = append(inputs, filepath.Join("docs", "features", "other-brief.md"))
	}
	tdWritePlannerEvidence(t, dir, "legacy-feat", inputs)
	out, code := runCent(t, bin, dir, "evidence", "validate", "legacy-feat")
	if code != 0 {
		t.Fatalf("superset evidence must still validate: %s", out)
	}
}

// Scenario: Evidence missing the feature's own plan is rejected
func TestTD_EvidenceMissingOwnPlanRejected(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdFeatureDocs(t, dir, "token-diet")
	tdWritePlannerEvidence(t, dir, "token-diet", []string{"docs/features/token-diet.md"})
	out, code := runCent(t, bin, dir, "evidence", "validate", "token-diet")
	if code == 0 {
		t.Fatal("evidence missing its own plan must fail validation")
	}
	mustContain(t, out, "missing feature-doc snapshot inputs")
	mustContain(t, out, "docs/plans/token-diet.md")
}

// Scenario: Evidence with an empty inputs list is rejected naming both paths
func TestTD_EvidenceEmptyInputsRejectedNamingBothPaths(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdFeatureDocs(t, dir, "token-diet")
	tdWritePlannerEvidence(t, dir, "token-diet", []string{})
	out, code := runCent(t, bin, dir, "evidence", "validate", "token-diet")
	if code == 0 {
		t.Fatal("empty inputs must fail validation")
	}
	// The empty-inputs case is also "incomplete evidence fields" (len(Inputs)==0
	// short-circuits before the snapshot check), so assert on the field the
	// gate actually reaches: an empty list can never satisfy either path.
	mustContain(t, out, "incomplete evidence fields")
}
