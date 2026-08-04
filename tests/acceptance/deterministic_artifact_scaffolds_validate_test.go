package acceptance_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init pre-fill lets big-thinker pass plan-snapshot validation with zero appends
func TestDAS_PreFillPassesSnapshotWithZeroAppends(t *testing.T) {
	dasChdir(t, "demo.md")
	if err := os.MkdirAll("docs/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/plans/demo.md", []byte("plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	// init pre-fills inputs; the snapshot rule must be satisfied by construction.
	skel := evidence.Skeleton("demo", orchestration.RoleBigThinker, "1.0.0")
	skel.Inputs = evidence.PlanInputs("demo", orchestration.RoleBigThinker)
	skel.Outputs = []string{"docs/plans/demo.md"}
	if err := evidence.WriteAtomic("demo", orchestration.RoleBigThinker, skel); err != nil {
		t.Fatal(err)
	}
	path := orchestration.JSONPath("demo", orchestration.RoleBigThinker)
	err := orchestration.ValidateEvidence(path, "demo", "plan", orchestration.RoleBigThinker, nil)
	if err != nil && strings.Contains(err.Error(), "missing feature-doc snapshot inputs") {
		t.Fatalf("snapshot rule failed despite pre-fill: %v", err)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Init pre-fill is idempotent under force re-run
func TestDAS_PreFillIdempotentUnderForce(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleBigThinker)
	first, _ := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	dasInit(t, "demo", orchestration.RoleBigThinker) // simulate --force re-run
	second, e := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	// The scenario asserts the pre-fill CONTENT is idempotent. generatedAt is a
	// second-granularity wall clock, so two writes straddling a tick differ by
	// design; comparing it made this a ~3%/run flake.
	if dasWithoutTimestamps(first) != dasWithoutTimestamps(second) {
		t.Fatalf("re-run not idempotent:\n%s\n---\n%s", first, second)
	}
	seen := map[string]int{}
	for _, in := range e.Inputs {
		seen[in]++
	}
	for in, n := range seen {
		if n != 1 {
			t.Fatalf("input %q duplicated %d times", in, n)
		}
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: A brief created after the first init does not change the pre-fill
// token-diet inverted this scenario: the pre-fill is construction-derived from
// the feature name, so it is invariant to what else lands in docs/features/.
func TestDAS_PreFillIgnoresLateSiblingBrief(t *testing.T) {
	dasChdir(t, "demo.md")
	dasInit(t, "demo", orchestration.RoleBigThinker)
	_, before := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	if err := os.WriteFile("docs/features/zzz-late.md", []byte("late"), 0o644); err != nil {
		t.Fatal(err)
	}
	dasInit(t, "demo", orchestration.RoleBigThinker) // --force after late brief
	_, after := dasReadJSON(t, "demo", orchestration.RoleBigThinker)
	if dasContains(after.Inputs, "docs/features/zzz-late.md") {
		t.Fatalf("late sibling brief leaked into the pre-fill: %v", after.Inputs)
	}
	if !reflect.DeepEqual(before.Inputs, after.Inputs) {
		t.Fatalf("pre-fill changed with an unrelated brief: %v -> %v", before.Inputs, after.Inputs)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Skeleton stays empty so repair and docs templates are not poisoned
func TestDAS_SkeletonStaysEmpty(t *testing.T) {
	skel := evidence.Skeleton("demo", orchestration.RoleBigThinker, "1.0.0")
	if len(skel.Inputs) != 0 {
		t.Fatalf("skeleton inputs not empty: %v", skel.Inputs)
	}
	data, err := evidence.SchemaSkeleton("", orchestration.RoleBigThinker, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "docs/plans/") {
		t.Fatalf("repair skeleton poisoned: %s", data)
	}
}
