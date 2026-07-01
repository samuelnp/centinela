package roadmap

import (
	"bytes"
	"testing"
)

// TestAdd_RejectionsByteIdentical drives every add rejection row and asserts the
// error substring plus a byte-identical roadmap.json (nothing is written).
func TestAdd_RejectionsByteIdentical(t *testing.T) {
	cases := []struct {
		name, slug, phase, substr string
		deps                      []string
	}{
		{"invalid slug", "Not_Kebab!", "Phase 1: Foundations", "invalid feature slug", nil},
		{"dup", "auth-service", "Phase 1: Foundations", "slug collision", nil},
		{"unknown phase", "new-widget", "Phase 9: Nonexistent", "unknown phase", nil},
		{"backlog target", "new-widget", "Backlog", "unknown phase", nil},
		{"baseline target", "new-widget", "Baseline", "unknown phase", nil},
		{"unknown dep", "new-widget", "Phase 1: Foundations", "depends on unknown feature", []string{"ghost-feature"}},
		{"self cycle", "new-widget", "Phase 1: Foundations", "roadmap dependency cycle detected", []string{"new-widget"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := crudWrite(t, crudBody)
			before := crudBytes(t, p)
			err := Add(p, AddRequest{Slug: c.slug, Phase: c.phase, DependsOn: c.deps})
			if err == nil {
				t.Fatalf("expected rejection for %s", c.name)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(c.substr)) {
				t.Fatalf("error %q must contain %q", err, c.substr)
			}
			if !bytes.Equal(before, crudBytes(t, p)) {
				t.Fatalf("%s must leave roadmap.json byte-identical", c.name)
			}
		})
	}
}

// TestAdd_DuplicateNamesOwningPhase reports the phase that already owns the slug.
func TestAdd_DuplicateNamesOwningPhase(t *testing.T) {
	p := crudWrite(t, crudBody)
	err := Add(p, AddRequest{Slug: "billing-api", Phase: "Phase 1: Foundations"})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("Phase 2: Growth")) {
		t.Fatalf("collision must name owning phase, got %v", err)
	}
}

// TestAdd_EmptyRoadmapUnknownPhase refuses to silently create a phase.
func TestAdd_EmptyRoadmapUnknownPhase(t *testing.T) {
	p := crudWrite(t, `{"phases":[]}`)
	before := crudBytes(t, p)
	err := Add(p, AddRequest{Slug: "new-widget", Phase: "Phase 1: Foundations"})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("unknown phase")) {
		t.Fatalf("empty roadmap add must error unknown phase, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("empty roadmap must remain exactly {\"phases\":[]}")
	}
}

// TestAdd_MissingFileSurfacesError leaves no file behind.
func TestAdd_MissingFileSurfacesError(t *testing.T) {
	p := crudWrite(t, crudBody) + ".absent"
	if err := Add(p, AddRequest{Slug: "new-widget", Phase: "Phase 1: Foundations"}); err == nil {
		t.Fatal("expected error for a missing roadmap.json")
	}
}
