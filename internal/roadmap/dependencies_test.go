package roadmap

import (
	"strings"
	"testing"
)

// Option B: cycle/unknown/self-dep detection moved off the analysis and onto
// roadmap.json deps; it is exercised against ValidateDependencies(*Roadmap).
func TestValidateDependenciesCycleAndUnknown(t *testing.T) {
	// Two-node cycle user→post→user.
	cyc := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "user", DependsOn: []string{"post"}},
		{Name: "post", DependsOn: []string{"user"}},
	}}}}
	if err := ValidateDependencies(cyc); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
	// Self-dependency A→A is a 1-node cycle.
	self := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "user", DependsOn: []string{"user"}},
	}}}}
	if err := ValidateDependencies(self); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected self-dep cycle error, got %v", err)
	}
	// Dependency on an unknown feature slug.
	unk := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "user", DependsOn: []string{"ghost"}},
	}}}}
	if err := ValidateDependencies(unk); err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected unknown dependency error, got %v", err)
	}
}

// Clean graphs (including nil) validate with no error.
func TestValidateDependenciesValidGraphs(t *testing.T) {
	if err := ValidateDependencies(nil); err != nil {
		t.Fatalf("nil roadmap must be dep-valid, got %v", err)
	}
	ok := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{
		{Name: "user"}, {Name: "post", DependsOn: []string{"user"}},
	}}}}
	if err := ValidateDependencies(ok); err != nil {
		t.Fatalf("valid graph must pass, got %v", err)
	}
}
