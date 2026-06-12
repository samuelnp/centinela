package workflow

import (
	"strings"
	"testing"
)

func TestNormalizeArchetype(t *testing.T) {
	cases := map[string]string{
		"":          ArchetypeCanonical, // empty defaults to canonical
		"hotfix":    "hotfix",           // known passthrough
		"  SPIKE  ": "spike",            // case + space normalized
		"bogus":     "bogus",            // unknown passes through for the validator
	}
	for in, want := range cases {
		if got := NormalizeArchetype(in); got != want {
			t.Fatalf("NormalizeArchetype(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateArchetype(t *testing.T) {
	for _, ok := range []string{"", ArchetypeCanonical, ArchetypeHotfix, ArchetypeRefactor, ArchetypeSpike} {
		if err := ValidateArchetype(ok); err != nil {
			t.Fatalf("ValidateArchetype(%q) must accept, got %v", ok, err)
		}
	}
	err := ValidateArchetype("nope")
	if err == nil {
		t.Fatal("unknown archetype must be rejected")
	}
	if !strings.Contains(err.Error(), "archetype") || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("error must name the field and value, got %q", err)
	}
}

func TestDisplayArchetype(t *testing.T) {
	if name, note := DisplayArchetype(nil); name != ArchetypeCanonical || note != "" {
		t.Fatalf("nil workflow → canonical, no note; got %q/%q", name, note)
	}
	for _, a := range []string{ArchetypeCanonical, ArchetypeHotfix, ArchetypeRefactor} {
		name, note := DisplayArchetype(&Workflow{Archetype: a})
		if name != a || note != "" {
			t.Fatalf("%s → %q/%q, want %q with no note", a, name, note, a)
		}
	}
	name, note := DisplayArchetype(&Workflow{Archetype: ArchetypeSpike})
	if name != ArchetypeSpike || !strings.Contains(note, "no ship gate") {
		t.Fatalf("spike must annotate no-ship-gate, got %q/%q", name, note)
	}
}
