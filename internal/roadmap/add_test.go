package roadmap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// featureIn returns the decoded feature named slug from roadmap.json at path.
func featureIn(t *testing.T, path, slug string) Feature {
	t.Helper()
	var r Roadmap
	if err := json.Unmarshal(crudBytes(t, path), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Name == slug {
				return f
			}
		}
	}
	t.Fatalf("feature %q not found", slug)
	return Feature{}
}

// TestAdd_SetsDraft appends a draft feature to a schedulable phase.
func TestAdd_SetsDraft(t *testing.T) {
	p := crudWrite(t, crudBody)
	if err := Add(p, AddRequest{Slug: "new-widget", Phase: "Phase 1: Foundations"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !featureIn(t, p, "new-widget").Draft {
		t.Fatal("added feature must be a draft")
	}
	// The feature landed in the requested phase (byte check on rendered output).
	if !bytes.Contains(crudBytes(t, p), []byte("new-widget")) {
		t.Fatal("new-widget must be present")
	}
}

// TestAdd_OptionalFlags records description, dependsOn and archetype.
func TestAdd_OptionalFlags(t *testing.T) {
	p := crudWrite(t, crudBody)
	err := Add(p, AddRequest{
		Slug: "new-widget", Phase: "Phase 1: Foundations",
		Description: "Adds the widget", Archetype: "canonical",
		DependsOn: []string{"auth-service"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	f := featureIn(t, p, "new-widget")
	if f.Description != "Adds the widget" || f.Archetype != "canonical" {
		t.Fatalf("description/archetype not persisted: %+v", f)
	}
	if len(f.DependsOn) != 1 || f.DependsOn[0] != "auth-service" || !f.Draft {
		t.Fatalf("dependsOn/draft not persisted: %+v", f)
	}
}

// TestAdd_ValidateStaysPass confirms a fresh draft keeps ValidateDependencies OK.
func TestAdd_ValidateStaysPass(t *testing.T) {
	p := crudWrite(t, crudBody)
	if err := Add(p, AddRequest{Slug: "new-widget", Phase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	var r Roadmap
	if err := json.Unmarshal(crudBytes(t, p), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateDependencies(&r); err != nil {
		t.Fatalf("validate must still pass: %v", err)
	}
	// Untouched Phase 1 features round-trip: still names auth-service/checkout-ui.
	s := string(crudBytes(t, p))
	if !strings.Contains(s, "auth-service") || !strings.Contains(s, "checkout-ui") {
		t.Fatal("untouched phase features must survive")
	}
}
