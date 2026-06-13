package workflow

import (
	"slices"
	"testing"
)

func TestArchetypeStepOrder_KnownAndUnknown(t *testing.T) {
	cases := map[string][]string{
		ArchetypeCanonical: {"plan", "code", "tests", "validate", "docs"},
		ArchetypeHotfix:    {"code", "tests", "validate"},
		ArchetypeRefactor:  {"plan", "code", "tests", "validate"},
		ArchetypeSpike:     {"plan", "code"},
	}
	for name, want := range cases {
		got, ok := ArchetypeStepOrder(name)
		if !ok || !slices.Equal(got, want) {
			t.Fatalf("ArchetypeStepOrder(%q) = %v,%v want %v,true", name, got, ok, want)
		}
	}
	if got, ok := ArchetypeStepOrder("mystery"); ok || got != nil {
		t.Fatalf("unknown archetype must return nil,false; got %v,%v", got, ok)
	}
}

// The canonical order must be a fresh clone, not an alias of DefaultStepOrder:
// mutating the returned slice must never corrupt the package-level default.
func TestArchetypeStepOrder_ReturnsClonesNotAliases(t *testing.T) {
	got, _ := ArchetypeStepOrder(ArchetypeCanonical)
	if len(got) == 0 {
		t.Fatal("canonical order must be non-empty")
	}
	got[0] = "MUTATED"
	if DefaultStepOrder[0] != "plan" {
		t.Fatalf("DefaultStepOrder was aliased and mutated: %v", DefaultStepOrder)
	}
	again, _ := ArchetypeStepOrder(ArchetypeCanonical)
	if again[0] != "plan" {
		t.Fatalf("second call returned a corrupted order: %v", again)
	}
}
