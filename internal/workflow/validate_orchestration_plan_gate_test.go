package workflow

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// A pinned workflow passes on the planner pair.
func TestValidateOrchestration_PinnedAcceptsPlannerEvidence(t *testing.T) {
	chdirTemp(t)
	seedPlanDocs(t, "fresh")
	mustSave(t, New("fresh"))
	writePlanEvidence(t, "fresh", orchestration.RolePlanner, "senior-engineer")
	if err := validateOrchestration("fresh", "plan", &config.Config{}); err != nil {
		t.Fatalf("planner evidence should satisfy a pinned workflow: %v", err)
	}
}

// The transition hole D2 closes: forged legacy files must not pass a fresh one.
func TestValidateOrchestration_PinnedRefusesForgedLegacyPair(t *testing.T) {
	chdirTemp(t)
	seedPlanDocs(t, "fresh")
	mustSave(t, New("fresh"))
	writePlanEvidence(t, "fresh", orchestration.RoleBigThinker, "feature-specialist")
	writePlanEvidence(t, "fresh", orchestration.RoleFeatureSpecial, "senior-engineer")
	err := validateOrchestration("fresh", "plan", &config.Config{})
	if err == nil {
		t.Fatal("a hand-authored legacy pair must not satisfy a planner-v1 workflow")
	}
	if !strings.Contains(err.Error(), "planner") {
		t.Fatalf("block message must name planner, got %v", err)
	}
	if strings.Contains(err.Error(), "predates") {
		t.Fatalf("a pinned workflow must not carry the legacy annotation: %v", err)
	}
}

// A legacy in-flight workflow keeps today's two-role gate verbatim.
func TestValidateOrchestration_LegacyAcceptsCompletePair(t *testing.T) {
	chdirTemp(t)
	seedPlanDocs(t, "old")
	mustSave(t, legacyWorkflow("old"))
	writePlanEvidence(t, "old", orchestration.RoleBigThinker, "feature-specialist")
	writePlanEvidence(t, "old", orchestration.RoleFeatureSpecial, "senior-engineer")
	if err := validateOrchestration("old", "plan", &config.Config{}); err != nil {
		t.Fatalf("legacy workflow must still pass on its own evidence: %v", err)
	}
}

func TestValidateOrchestration_LegacyRejectsPartialPair(t *testing.T) {
	chdirTemp(t)
	seedPlanDocs(t, "old")
	mustSave(t, legacyWorkflow("old"))
	writePlanEvidence(t, "old", orchestration.RoleBigThinker, "feature-specialist")
	err := validateOrchestration("old", "plan", &config.Config{})
	if err == nil {
		t.Fatal("a partial legacy set must still fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, "feature-specialist") {
		t.Fatalf("message must name the missing half of the legacy pair: %v", err)
	}
	if !strings.Contains(msg, "predates") || !strings.Contains(msg, PlanContractUnified) {
		t.Fatalf("message must carry the contract annotation: %v", err)
	}
}

// A legacy workflow is never retro-gated on planner evidence it never wrote.
func TestValidateOrchestration_LegacyNotRetroGatedOnPlanner(t *testing.T) {
	chdirTemp(t)
	seedPlanDocs(t, "old")
	mustSave(t, legacyWorkflow("old"))
	writePlanEvidence(t, "old", orchestration.RoleBigThinker, "feature-specialist")
	writePlanEvidence(t, "old", orchestration.RoleFeatureSpecial, "senior-engineer")
	if err := validateOrchestration("old", "plan", &config.Config{}); err != nil {
		t.Fatalf("legacy pass must not require planner evidence: %v", err)
	}
}

// Non-strict profiles skip orchestration evidence entirely, both contracts.
func TestValidateOrchestration_NonStrictSkipsPlanGate(t *testing.T) {
	chdirTemp(t)
	wf := NewWithOrder("guided", DefaultStepOrder, config.ProfileGuided)
	mustSave(t, wf)
	if err := validateOrchestration("guided", "plan", &config.Config{}); err != nil {
		t.Fatalf("guided profile must not require plan evidence: %v", err)
	}
}
