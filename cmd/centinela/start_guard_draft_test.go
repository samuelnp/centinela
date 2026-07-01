package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestDraftStartError names the feature and the finalize path.
func TestDraftStartError(t *testing.T) {
	err := draftStartError("my-widget")
	if err == nil || !strings.Contains(err.Error(), "draft") ||
		!strings.Contains(err.Error(), "my-widget") {
		t.Fatalf("draftStartError must mention draft + the slug, got %v", err)
	}
}

// TestResolveArchetypeOrder_RefusesDraft mirrors the Backlog refusal for drafts.
func TestResolveArchetypeOrder_RefusesDraft(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile,
		`{"phases":[{"name":"Phase 1","features":[{"name":"the-draft","draft":true}]}]}`)
	_, _, err := resolveArchetypeOrder("the-draft", "")
	if err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("start must refuse a draft, got %v", err)
	}
}

// TestResolveArchetypeOrder_NonDraftFlag resolves a non-draft via the flag path.
func TestResolveArchetypeOrder_NonDraftFlag(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile,
		`{"phases":[{"name":"Phase 1","features":[{"name":"auth-service"}]}]}`)
	order, name, err := resolveArchetypeOrder("auth-service", "canonical")
	if err != nil || name != "canonical" || len(order) == 0 {
		t.Fatalf("non-draft flag path: order=%v name=%q err=%v", order, name, err)
	}
}
