package roadmap

import (
	"strings"
	"testing"
)

// TestValidateDependencies_DependOnDraftAllowed accepts a non-draft that depends
// on a draft — a draft is a real, dependable feature.
func TestValidateDependencies_DependOnDraftAllowed(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "d", Draft: true},
		{Name: "a", DependsOn: []string{"d"}},
	}}}}
	if err := ValidateDependencies(r); err != nil {
		t.Fatalf("depending on a draft must validate: %v", err)
	}
}

// TestValidateDependencies_SelfDepIsCycle treats a self-dependency as a cycle,
// not an unknown-dependency error.
func TestValidateDependencies_SelfDepIsCycle(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "a", DependsOn: []string{"a"}},
	}}}}
	err := ValidateDependencies(r)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("self-dependency must be a cycle, got %v", err)
	}
}

// TestValidateDependencies_UnknownDep still rejects a truly-unknown target.
func TestValidateDependencies_UnknownDep(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "a", DependsOn: []string{"ghost"}},
	}}}}
	err := ValidateDependencies(r)
	if err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("unknown dep must error, got %v", err)
	}
}
