package roadmap

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestDefer_InvalidSlug rejected before any write.
func TestDefer_InvalidSlug(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	before, _ := os.ReadFile(p)
	err := Defer(p, DeferOptions{Slug: "Bad Slug!", Summary: "Something"})
	if err == nil {
		t.Fatal("invalid slug must be rejected")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on invalid slug")
	}
}

// TestDefer_SlugCollisionInBacklog refused when slug already in Backlog (duplicate-entry regression).
func TestDefer_SlugCollisionInBacklog(t *testing.T) {
	src := `{"phases":[{"name":"Backlog","features":[{"name":"already-here","summary":"x","deferredAt":"t"}]}]}`
	_, p := deferSetup(t, src)
	before, _ := os.ReadFile(p)
	err := Defer(p, DeferOptions{Slug: "already-here", Summary: "again"})
	if err == nil {
		t.Fatal("duplicate Backlog slug must be refused")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on collision")
	}
}

// TestDefer_SlugCollisionInNonBacklog refused when slug exists elsewhere (regression).
func TestDefer_SlugCollisionInNonBacklog(t *testing.T) {
	src := `{"phases":[{"name":"Phase 5","features":[{"name":"shipped-feature"}]}]}`
	_, p := deferSetup(t, src)
	before, _ := os.ReadFile(p)
	err := Defer(p, DeferOptions{Slug: "shipped-feature", Summary: "but it shipped"})
	if err == nil {
		t.Fatal("collision in non-Backlog phase must be refused")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on non-Backlog collision")
	}
}

// TestDefer_NoSourceField writes entry without source key when Source is nil.
func TestDefer_NoSourceField(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	if err := Defer(p, DeferOptions{Slug: "no-src", Summary: "no source provided"}); err != nil {
		t.Fatalf("Defer without source: %v", err)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), `"source"`) {
		t.Error("source key must be absent when Source is nil")
	}
}

// TestDefer_AppendToExistingBacklog adds second entry preserving first.
func TestDefer_AppendToExistingBacklog(t *testing.T) {
	src := `{"phases":[{"name":"Backlog","features":[{"name":"first","summary":"f1","deferredAt":"t"}]}]}`
	_, p := deferSetup(t, src)
	if err := Defer(p, DeferOptions{Slug: "second", Summary: "s2"}); err != nil {
		t.Fatalf("Defer second entry: %v", err)
	}
	data, _ := os.ReadFile(p)
	s := string(data)
	if !strings.Contains(s, "first") {
		t.Error("first entry must be preserved")
	}
	if !strings.Contains(s, "second") {
		t.Error("second entry must be appended")
	}
}
