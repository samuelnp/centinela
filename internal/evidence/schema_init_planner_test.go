package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestSkeleton_PlannerStepAndHandoff(t *testing.T) {
	skel := Skeleton("f", orchestration.RolePlanner, "0.0.0")
	if skel.Step != "plan" {
		t.Fatalf("planner step = %q, want plan", skel.Step)
	}
	if skel.HandoffTo != string(orchestration.RoleSeniorEngineer) {
		t.Fatalf("planner handoffTo = %q, want senior-engineer", skel.HandoffTo)
	}
	if skel.Role != "planner" {
		t.Fatalf("planner role = %q", skel.Role)
	}
}

// Legacy chaining is untouched: big-thinker still hands off to feature-specialist.
func TestSkeleton_LegacyPlanChainUnchanged(t *testing.T) {
	if got := Skeleton("f", orchestration.RoleBigThinker, "0").HandoffTo; got != "feature-specialist" {
		t.Fatalf("big-thinker handoffTo = %q, want feature-specialist", got)
	}
	if got := Skeleton("f", orchestration.RoleFeatureSpecial, "0").HandoffTo; got != "senior-engineer" {
		t.Fatalf("feature-specialist handoffTo = %q, want senior-engineer", got)
	}
	if got := Skeleton("f", orchestration.RoleBigThinker, "0").Step; got != "plan" {
		t.Fatalf("big-thinker step = %q, want plan", got)
	}
}

func TestAllRoles_IncludesPlannerAndKeepsLegacy(t *testing.T) {
	for _, slug := range []string{"planner", "big-thinker", "feature-specialist"} {
		if !IsKnownRole(Role(slug)) {
			t.Errorf("AllRoles must still accept %q", slug)
		}
	}
	if _, err := ParseRole("planner"); err != nil {
		t.Fatalf("ParseRole(planner): %v", err)
	}
}

func TestPlanInputs_PlannerInheritsSnapshotRule(t *testing.T) {
	got := PlanInputs("f", orchestration.RolePlanner)
	want := orchestration.RequiredPlanInputs("f")
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("planner PlanInputs = %v, want %v", got, want)
	}
	if PlanInputs("f", orchestration.RoleSeniorEngineer) != nil {
		t.Fatal("non-plan roles must derive no inputs")
	}
}

// The companion skeleton carries the ordered union of both lenses.
func TestCompanionSkeleton_PlannerUnionHeadersInOrder(t *testing.T) {
	body, ok := companionSkeleton("f", orchestration.RolePlanner)
	if !ok {
		t.Fatal("planner must have a companion skeleton")
	}
	ordered := []string{"## Problem", "## Scope", "## Dependencies & Assumptions",
		"## Risks", "## Rollout", "## Behavior Summary",
		"## Acceptance Criteria (Gherkin)", "## UX States", "## Edge Cases",
		"## Out-of-Scope", "## Handoff"}
	prev := -1
	for _, h := range ordered {
		i := strings.Index(body, h)
		if i < 0 {
			t.Fatalf("planner skeleton missing %q", h)
		}
		if i <= prev {
			t.Fatalf("planner skeleton header %q is out of order", h)
		}
		prev = i
	}
}
