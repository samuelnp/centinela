package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// editBody: checkout-ui carries description/archetype/dependsOn so field-only
// edits can prove the untouched fields survive; Phase 2 holds an untouched sibling.
const editBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","description":"Original description","archetype":"canonical",` +
	`"dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]}]}`

// TestEdit_OnlyProvidedFieldChanges edits description; archetype/deps and the
// untouched Phase 2 stay intact.
func TestEdit_OnlyProvidedFieldChanges(t *testing.T) {
	p, before := canonRoadmap(t, editBody)
	if err := Edit(p, EditRequest{Slug: "checkout-ui", Description: "Updated description"}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	f := featureIn(t, p, "checkout-ui")
	if f.Description != "Updated description" || f.Archetype != "canonical" {
		t.Fatalf("only description should change: %+v", f)
	}
	if len(f.DependsOn) != 1 || f.DependsOn[0] != "auth-service" {
		t.Fatalf("deps must stay intact (unchanged, not cleared): %+v", f.DependsOn)
	}
	if !bytes.Contains(crudBytes(t, p), phaseSlice(t, before, "Phase 2: Growth")) {
		t.Fatal("untouched Phase 2 must be byte-identical")
	}
}

// TestEdit_DependsOnUnchangedVsClear proves the SetDeps sentinel: omitted keeps
// deps, an explicit empty set clears them.
func TestEdit_DependsOnUnchangedVsClear(t *testing.T) {
	p, _ := canonRoadmap(t, editBody)
	if err := Edit(p, EditRequest{Slug: "checkout-ui", Description: "x"}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := featureIn(t, p, "checkout-ui").DependsOn; len(got) != 1 {
		t.Fatalf("omitted --depends-on must not clear: %+v", got)
	}
	if err := Edit(p, EditRequest{Slug: "checkout-ui", SetDeps: true}); err != nil {
		t.Fatalf("Edit clear: %v", err)
	}
	if got := featureIn(t, p, "checkout-ui").DependsOn; len(got) != 0 {
		t.Fatalf("explicit empty --depends-on must clear: %+v", got)
	}
}

// TestEdit_DependsOnReplace swaps the dependency list to a new valid target.
func TestEdit_DependsOnReplace(t *testing.T) {
	p, _ := canonRoadmap(t, editBody)
	err := Edit(p, EditRequest{Slug: "checkout-ui", DependsOn: []string{"billing-api"}, SetDeps: true})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := featureIn(t, p, "checkout-ui").DependsOn; len(got) != 1 || got[0] != "billing-api" {
		t.Fatalf("deps must be replaced: %+v", got)
	}
}

// TestEdit_NotFoundByteIdentical rejects an unknown slug and writes nothing.
func TestEdit_NotFoundByteIdentical(t *testing.T) {
	p, before := canonRoadmap(t, editBody)
	err := Edit(p, EditRequest{Slug: "ghost-feature", Description: "x"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unknown slug must error not-found, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, p)) {
		t.Fatal("rejected edit must leave roadmap.json byte-identical")
	}
}
