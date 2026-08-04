// Integration coverage for docs/plans/evidence-schema-skeleton-legacy-handoff.md.
// Three packages meet here: internal/evidence renders the skeleton,
// internal/workflow derives and re-checks the successor, internal/orchestration
// decides which roles a contract requires. The invariant under test is that a
// value the CLI PRINTS is a value the completion gate ACCEPTS — for every
// archetype and contract pin, from the worktree root and from any depth in it.
package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

type schemaFixture struct {
	name, feature, order, pins, brief, role, step, want string
}

// schemaFixtures spans the contract shapes whose successors differ: an internal
// canonical feature (docs step requires nobody), a user-facing one (same-step
// ux-ui hop), the two archetypes that drop steps, and a legacy-pinned workflow
// with neither contract field, whose validate step is occupied by the
// validation-specialist rather than the gatekeeper.
func schemaFixtures() []schemaFixture {
	canonical := `"plan","code","tests","validate","docs"`
	pins := `"validateContract":"adversarial-v1","planContract":"unified-v1",`
	return []schemaFixture{
		{"canonical internal", "demo-internal", canonical, pins, "# demo\n", "gatekeeper", "validate", "complete"},
		{"canonical user-facing", "demo-uxfacing", canonical, pins, "surface: user-facing\n", "senior-engineer", "code", "ux-ui-specialist"},
		{"hotfix drops plan and docs", "demo-hotfix", `"code","tests","validate"`, pins, "# demo\n", "gatekeeper", "validate", "complete"},
		{"spike has no validate step", "demo-spike", `"plan","code"`, pins, "# demo\n", "planner", "plan", "senior-engineer"},
		{"legacy pins, user-facing", "demo-legacy", canonical, "", "surface: user-facing\n", "validation-specialist", "validate", "documentation-specialist"},
	}
}

// schemaWorktreeFixture lays out <tmp>/.worktrees/<feature> with workflow state,
// a brief and a nested package dir, and returns the root and that nested dir.
func schemaWorktreeFixture(t *testing.T, f schemaFixture) (root, deep string) {
	t.Helper()
	base := t.TempDir()
	if r, err := filepath.EvalSymlinks(base); err == nil {
		base = r
	}
	root = filepath.Join(base, ".worktrees", f.feature)
	deep = filepath.Join(root, "internal", "evidence")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, root, ".workflow/"+f.feature+".json",
		`{"feature":"`+f.feature+`","currentStep":"`+f.step+`","steps":{},`+
			f.pins+`"stepOrder":[`+f.order+`]}`)
	mustWrite(t, root, "docs/features/"+f.feature+".md", f.brief)
	return root, deep
}

// schemaHandoffFrom reproduces the command's two-line contract — resolve from
// the CWD, derive from the ROOT the resolver hands back — and returns the
// printed handoffTo. A root that does not point at the state is exactly the
// defect this asserts against: derivation would silently answer from the legacy
// static chain instead.
func schemaHandoffFrom(t *testing.T, cwd string, role evidence.Role) string {
	t.Helper()
	feature, root := evidence.ResolveActiveFeature(cwd)
	if root != "" {
		t.Chdir(root)
	}
	data, err := evidence.SchemaSkeleton(feature, role, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		HandoffTo string `json:"handoffTo"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("skeleton is not valid JSON: %v\n%s", err, data)
	}
	return got.HandoffTo
}

// TestSchemaHandoffRoundTripsThroughTheGate: for every contract shape, the value
// printed from the root and from a subdirectory is the same, is the expected
// successor, and survives being written back as evidence — CheckHandoffTo and
// the evidence validator both accept it.
func TestSchemaHandoffRoundTripsThroughTheGate(t *testing.T) {
	for _, f := range schemaFixtures() {
		t.Run(f.name, func(t *testing.T) {
			root, deep := schemaWorktreeFixture(t, f)
			role := evidence.Role(f.role)
			fromRoot := schemaHandoffFrom(t, root, role)
			fromDeep := schemaHandoffFrom(t, deep, role)
			if fromRoot != f.want || fromDeep != f.want {
				t.Fatalf("handoffTo root %q / subdir %q, want %q", fromRoot, fromDeep, f.want)
			}
			t.Chdir(root)
			if err := workflow.CheckHandoffTo(f.feature, f.step, orchestration.Role(f.role), fromDeep); err != nil {
				t.Fatalf("the gate refuses the value the CLI printed: %v", err)
			}
			assertEvidenceAccepts(t, f, role, fromDeep)
		})
	}
}

// assertEvidenceAccepts writes the printed value back with the real evidence
// writer and runs the validator over it, so the round trip goes through the
// same file the completion gate reads rather than a direct function call.
func assertEvidenceAccepts(t *testing.T, f schemaFixture, role evidence.Role, handoff string) {
	t.Helper()
	r := evidence.Skeleton(f.feature, role, "1.0.0")
	if err := evidence.SetField(r, "handoffTo", handoff); err != nil {
		t.Fatal(err)
	}
	r.Inputs = []string{"docs/plans/" + f.feature + ".md"}
	r.Outputs = []string{"docs/features/" + f.feature + ".md"}
	r.EdgeCases = []string{"round-trip fixture"}
	if err := evidence.WriteAtomic(f.feature, role, r); err != nil {
		t.Fatal(err)
	}
	for _, h := range evidence.ValidateFeature(f.feature, nil) {
		if h.Field == "handoffTo" {
			t.Fatalf("validator rejects the printed value: %+v", h)
		}
	}
}
