// Acceptance: specs/workflow-archetypes.feature
package acceptance_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: Archetype and enforcement profile are independent
func TestWA_ArchetypeIndependentOfProfile(t *testing.T) {
	order, _ := workflow.ArchetypeStepOrder(workflow.ArchetypeSpike)
	wf := workflow.NewWithOrder("feat", order, config.ProfileStrict)
	if !slices.Equal(wf.StepOrder, []string{"plan", "code"}) {
		t.Fatalf("step order must be the spike order, got %v", wf.StepOrder)
	}
	if wf.EnforcementProfile != config.ProfileStrict {
		t.Fatalf("enforcement profile must be strict, got %q", wf.EnforcementProfile)
	}
}

// Scenario: An unknown archetype value is rejected
func TestWA_UnknownArchetypeRejected(t *testing.T) {
	err := workflow.ValidateArchetype("warpdrive")
	if err == nil {
		t.Fatal("an unsupported archetype must fail validation")
	}
	if !strings.Contains(err.Error(), "archetype") || !strings.Contains(err.Error(), "warpdrive") {
		t.Fatalf("error must name the archetype, got %q", err)
	}
}

// Scenario: The status output shows the active archetype
func TestWA_StatusShowsArchetype(t *testing.T) {
	wf := &workflow.Workflow{Feature: "f", CurrentStep: "code", Archetype: workflow.ArchetypeSpike}
	out := ui.RenderStatus(wf)
	if !strings.Contains(out, "spike") {
		t.Fatalf("status output must name the spike archetype, got:\n%s", out)
	}
}
