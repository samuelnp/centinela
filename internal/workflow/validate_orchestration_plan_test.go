package workflow

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func roleSlugs(roles []orchestration.Role) []string {
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, string(r))
	}
	return out
}

func TestRequiredEvidenceRoles_PinnedPlanDemandsPlannerOnly(t *testing.T) {
	chdirTemp(t)
	mustSave(t, New("fresh"))
	got := roleSlugs(RequiredEvidenceRoles("fresh", "plan"))
	if len(got) != 1 || got[0] != "planner" {
		t.Fatalf("pinned plan roles = %v, want [planner]", got)
	}
}

func TestRequiredEvidenceRoles_UnpinnedPlanDemandsLegacyPair(t *testing.T) {
	chdirTemp(t)
	mustSave(t, legacyWorkflow("old"))
	got := roleSlugs(RequiredEvidenceRoles("old", "plan"))
	want := []string{"big-thinker", "feature-specialist"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unpinned plan roles = %v, want %v", got, want)
	}
}

func TestRequiredEvidenceRoles_MissingStateFallsBackToLegacyPair(t *testing.T) {
	chdirTemp(t)
	got := roleSlugs(RequiredEvidenceRoles("ghost", "plan"))
	if len(got) != 2 {
		t.Fatalf("missing-state plan roles = %v, want the legacy pair", got)
	}
}

func TestRequiredEvidenceRoles_NonPlanStepsUnaffected(t *testing.T) {
	chdirTemp(t)
	mustSave(t, legacyWorkflow("old"))
	got := roleSlugs(RequiredEvidenceRoles("old", "tests"))
	if len(got) != 1 || got[0] != "qa-senior" {
		t.Fatalf("tests roles = %v, want [qa-senior]", got)
	}
}

func TestAnnotatePlanContract_ExplainsLegacyRequirement(t *testing.T) {
	chdirTemp(t)
	mustSave(t, legacyWorkflow("old"))
	err := annotatePlanContract("old", "plan", errSentinel)
	if err == nil || !strings.Contains(err.Error(), "predates") {
		t.Fatalf("expected contract annotation, got %v", err)
	}
	if !strings.Contains(err.Error(), "big-thinker") ||
		!strings.Contains(err.Error(), "feature-specialist") {
		t.Fatalf("annotation must name this workflow's required set, got %v", err)
	}
	if strings.Contains(err.Error(), "planner-v1\", so its plan step requires the complete legacy planner") {
		t.Fatalf("annotation must not offer planner as an alternative: %v", err)
	}
}

func TestAnnotatePlanContract_SilentForPinnedAndNonPlan(t *testing.T) {
	chdirTemp(t)
	mustSave(t, New("fresh"))
	if got := annotatePlanContract("fresh", "plan", errSentinel); got != errSentinel {
		t.Fatalf("pinned workflow must not be annotated, got %v", got)
	}
	if got := annotatePlanContract("fresh", "code", errSentinel); got != errSentinel {
		t.Fatalf("non-plan step must not be annotated, got %v", got)
	}
	if got := annotatePlanContract("fresh", "plan", nil); got != nil {
		t.Fatalf("a passing gate must stay nil, got %v", got)
	}
}
