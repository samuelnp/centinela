package roadmap

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFeatureArchetype_Parse(t *testing.T) {
	var r Roadmap
	raw := `{"phases":[{"name":"P0","features":[{"name":"fix","archetype":"hotfix"}]}]}`
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := FeatureArchetype(&r, "fix"); got != "hotfix" {
		t.Fatalf("parsed archetype = %q, want hotfix", got)
	}
}

func TestFeatureArchetype_Accessor(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "a", Archetype: "spike"}, {Name: "b"},
	}}}}
	if got := FeatureArchetype(r, "a"); got != "spike" {
		t.Fatalf("a → %q, want spike", got)
	}
	if got := FeatureArchetype(r, "b"); got != "" {
		t.Fatalf("b has no archetype, got %q", got)
	}
	if got := FeatureArchetype(r, "missing"); got != "" {
		t.Fatalf("missing feature → empty, got %q", got)
	}
	if got := FeatureArchetype(nil, "a"); got != "" {
		t.Fatalf("nil roadmap → empty, got %q", got)
	}
}

// A roadmap that pins an unsupported archetype must fail validation on load,
// naming the offending feature, so a bad roadmap fails fast.
func TestValidateDependencies_RejectsUnknownArchetype(t *testing.T) {
	bad := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "weird", Archetype: "turbo"},
	}}}}
	err := ValidateDependencies(bad)
	if err == nil || !strings.Contains(err.Error(), "weird") {
		t.Fatalf("expected error naming feature weird, got %v", err)
	}
	good := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "ok", Archetype: "refactor"},
	}}}}
	if err := ValidateDependencies(good); err != nil {
		t.Fatalf("valid archetype must pass, got %v", err)
	}
}
