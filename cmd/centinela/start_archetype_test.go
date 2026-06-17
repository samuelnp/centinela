package main

import (
	"os"
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// archetypeOrderByName validates then resolves a named archetype to its order.
func TestArchetypeOrderByName(t *testing.T) {
	order, name, err := archetypeOrderByName("HOTFIX")
	if err != nil || name != "hotfix" || !slices.Equal(order, []string{"code", "tests", "validate"}) {
		t.Fatalf("hotfix → %v/%q/%v", order, name, err)
	}
	if _, _, err := archetypeOrderByName("bogus"); err == nil {
		t.Fatal("unknown archetype must be rejected")
	}
}

// Precedence tier 1: an explicit --archetype flag wins over everything, even a
// roadmap that pins a different archetype for the same feature.
func TestResolveArchetypeOrder_FlagOverridesRoadmap(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P0", Features: []roadmap.Feature{
		{Name: "feat", Archetype: "refactor"},
	}}}}
	roadmap.Save(r) //nolint:errcheck
	order, name, err := resolveArchetypeOrder("feat", "hotfix")
	if err != nil || name != "hotfix" || !slices.Equal(order, []string{"code", "tests", "validate"}) {
		t.Fatalf("flag must override roadmap: %v/%q/%v", order, name, err)
	}
}

// Precedence tier 2: with no flag, the roadmap Feature archetype is used.
func TestResolveArchetypeOrder_RoadmapWhenNoFlag(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P0", Features: []roadmap.Feature{
		{Name: "feat", Archetype: "spike"},
	}}}}
	roadmap.Save(r) //nolint:errcheck
	order, name, err := resolveArchetypeOrder("feat", "")
	if err != nil || name != "spike" || !slices.Equal(order, []string{"plan", "code"}) {
		t.Fatalf("roadmap archetype must apply: %v/%q/%v", order, name, err)
	}
}

// Precedence tier 3: no flag and no roadmap archetype falls through to the
// canonical/bootstrap order from workflowOrderForFeature.
func TestResolveArchetypeOrder_FallsThroughToCanonical(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	order, name, err := resolveArchetypeOrder("feat", "")
	if err != nil || name != workflow.ArchetypeCanonical {
		t.Fatalf("expected canonical fallthrough, got %q/%v", name, err)
	}
	if !slices.Equal(order, workflow.DefaultStepOrder) {
		t.Fatalf("expected default order, got %v", order)
	}
}
