// Acceptance: specs/workflow-archetypes.feature
package acceptance_test

import (
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func waOrder(t *testing.T, name string) []string {
	t.Helper()
	order, ok := workflow.ArchetypeStepOrder(name)
	if !ok {
		t.Fatalf("archetype %q must be known", name)
	}
	return order
}

// Scenario: The hotfix archetype resolves to a code-tests-validate order
func TestWA_HotfixOrder(t *testing.T) {
	order := waOrder(t, workflow.ArchetypeHotfix)
	if !slices.Equal(order, []string{"code", "tests", "validate"}) {
		t.Fatalf("hotfix order = %v", order)
	}
	if slices.Contains(order, "plan") || slices.Contains(order, "docs") {
		t.Fatalf("hotfix must omit plan and docs, got %v", order)
	}
}

// Scenario: The refactor archetype resolves to a plan-code-tests-validate order
func TestWA_RefactorOrder(t *testing.T) {
	order := waOrder(t, workflow.ArchetypeRefactor)
	if !slices.Equal(order, []string{"plan", "code", "tests", "validate"}) {
		t.Fatalf("refactor order = %v", order)
	}
	if slices.Contains(order, "docs") {
		t.Fatalf("refactor must omit docs, got %v", order)
	}
}

// Scenario: The spike archetype resolves to a plan-code order with no validate step
func TestWA_SpikeOrder(t *testing.T) {
	order := waOrder(t, workflow.ArchetypeSpike)
	if !slices.Equal(order, []string{"plan", "code"}) {
		t.Fatalf("spike order = %v", order)
	}
	if slices.Contains(order, "validate") {
		t.Fatalf("spike must contain no validate step, got %v", order)
	}
}

// Scenario: The default archetype is the canonical five-step order
func TestWA_CanonicalDefault(t *testing.T) {
	if workflow.NormalizeArchetype("") != workflow.ArchetypeCanonical {
		t.Fatal("empty archetype must default to canonical")
	}
	order := waOrder(t, workflow.ArchetypeCanonical)
	if !slices.Equal(order, []string{"plan", "code", "tests", "validate", "docs"}) {
		t.Fatalf("canonical order = %v", order)
	}
}
