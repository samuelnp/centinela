package workflow

import (
	"slices"
	"testing"
)

// The feature's core safety claim: a spike skips the ship gate ONLY by having no
// validate step — never via a bypass branch. The ship gate (complete.go) fires
// iff the resolved order contains "validate". This test pins that property for
// every archetype, so the gate's correctness rests on observable step orders.
func TestArchetypeSafety_ShipGateKeyedOnValidatePresence(t *testing.T) {
	cases := []struct {
		archetype     string
		wantOrder     []string
		wantValidates bool
	}{
		{ArchetypeSpike, []string{"plan", "code"}, false},
		{ArchetypeHotfix, []string{"code", "tests", "validate"}, true},
		{ArchetypeRefactor, []string{"plan", "code", "tests", "validate"}, true},
		{ArchetypeCanonical, []string{"plan", "code", "tests", "validate", "docs"}, true},
	}
	for _, c := range cases {
		order, ok := ArchetypeStepOrder(c.archetype)
		if !ok {
			t.Fatalf("%s: archetype must be known", c.archetype)
		}
		if !slices.Equal(order, c.wantOrder) {
			t.Fatalf("%s: order = %v, want %v", c.archetype, order, c.wantOrder)
		}
		hasValidate := slices.Contains(order, "validate")
		if hasValidate != c.wantValidates {
			t.Fatalf("%s: contains validate = %v, want %v (order %v)",
				c.archetype, hasValidate, c.wantValidates, order)
		}
	}
}

// Pinned exactly: spike's order is [plan, code] and contains no "validate"
// element, so the step-keyed ship gate can never fire for it.
func TestArchetypeSafety_SpikeOrderIsPlanCodeNoValidate(t *testing.T) {
	order, ok := ArchetypeStepOrder(ArchetypeSpike)
	if !ok || !slices.Equal(order, []string{"plan", "code"}) {
		t.Fatalf("spike order must be exactly [plan code], got %v (ok=%v)", order, ok)
	}
	if slices.Contains(order, "validate") {
		t.Fatalf("spike order must contain no validate step, got %v", order)
	}
}
