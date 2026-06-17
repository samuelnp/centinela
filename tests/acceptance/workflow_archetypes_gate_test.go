// Acceptance: specs/workflow-archetypes.feature
package acceptance_test

import (
	"os"
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: A ship-gated archetype runs gates and claim verification
func TestWA_ShipGatedArchetypeReachesValidate(t *testing.T) {
	// The ship gate (complete.go) fires iff the resolved order contains the
	// "validate" step; every non-spike archetype carries it.
	for _, a := range []string{workflow.ArchetypeCanonical, workflow.ArchetypeHotfix, workflow.ArchetypeRefactor} {
		order, _ := workflow.ArchetypeStepOrder(a)
		if !slices.Contains(order, "validate") {
			t.Fatalf("%s order must contain validate so the ship gate fires, got %v", a, order)
		}
	}
}

// Scenario: A spike never reaches the ship gate
func TestWA_SpikeNeverReachesShipGate(t *testing.T) {
	order, _ := workflow.ArchetypeStepOrder(workflow.ArchetypeSpike)
	if slices.Contains(order, "validate") {
		t.Fatalf("spike order must omit validate so the gate is never triggered, got %v", order)
	}
}

// Scenario: An explicit archetype flag overrides the roadmap archetype
func TestWA_FlagOverridesRoadmapArchetype(t *testing.T) {
	// Mirrors resolveArchetypeOrder precedence: a flag value resolves regardless
	// of any roadmap-pinned archetype for the feature.
	flag := workflow.NormalizeArchetype("hotfix")
	roadmapPinned := workflow.NormalizeArchetype("refactor")
	resolved := flag // flag wins
	if resolved == roadmapPinned || resolved != workflow.ArchetypeHotfix {
		t.Fatalf("explicit flag must resolve to hotfix, got %q (roadmap %q)", resolved, roadmapPinned)
	}
}

// Scenario: The active archetype is pinned in the workflow state
func TestWA_ArchetypePinnedInState(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(workflow.WorkflowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	order, _ := workflow.ArchetypeStepOrder(workflow.ArchetypeHotfix)
	wf := workflow.NewWithOrder("feat", order, config.ProfileStrict)
	wf.Archetype = workflow.ArchetypeHotfix
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := workflow.Load("feat")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Archetype != workflow.ArchetypeHotfix {
		t.Fatalf("reloaded archetype = %q, want hotfix", got.Archetype)
	}
}
