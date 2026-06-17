package workflow

import (
	"os"
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// The pinned archetype must survive a Save/Load round-trip through the .workflow
// JSON, so status and any later resolution see the track chosen at start.
func TestArchetype_StateRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(WorkflowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	order, _ := ArchetypeStepOrder(ArchetypeHotfix)
	wf := NewWithOrder("feat", order, config.ProfileStrict)
	wf.Archetype = ArchetypeHotfix
	if err := Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load("feat")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Archetype != ArchetypeHotfix {
		t.Fatalf("archetype not persisted: %q", got.Archetype)
	}
	if !slices.Equal(got.StepOrder, []string{"code", "tests", "validate"}) {
		t.Fatalf("order not persisted: %v", got.StepOrder)
	}
}

// Archetype (sequence) and enforcement profile (strictness) are orthogonal: a
// spike order combined with the strict profile yields the spike order AND the
// strict profile, with no shared code coupling the two axes.
func TestArchetype_OrthogonalToProfile(t *testing.T) {
	order, _ := ArchetypeStepOrder(ArchetypeSpike)
	wf := NewWithOrder("feat", order, config.ProfileStrict)
	if !slices.Equal(wf.StepOrder, []string{"plan", "code"}) {
		t.Fatalf("spike order expected, got %v", wf.StepOrder)
	}
	if wf.EnforcementProfile != config.ProfileStrict {
		t.Fatalf("strict profile expected, got %q", wf.EnforcementProfile)
	}
}
