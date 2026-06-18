package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

func TestRenderInventorySummary_FullInventory(t *testing.T) {
	inv := analyze.Inventory{
		SchemaVersion:   1,
		PrimaryLanguage: "Go",
		Manifests: []analyze.Manifest{
			{Kind: "make", Path: "Makefile", Build: "make build", Test: "make test"},
		},
		Locales:  []string{"en", "es"},
		Packages: []string{"a", "b", "c"},
		Graph:    analyze.DependencyGraph{Kind: "go-packages", Edges: []analyze.Edge{{From: "a", To: "b"}}},
	}
	out := RenderInventorySummary(inv)
	for _, want := range []string{
		"schema v1", "primary language: Go", "make build", "make test",
		"locales: 2", "packages: 3", "graph edges: 1 (go-packages)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("summary missing %q:\n%s", want, out)
		}
	}
}

func TestRenderInventorySummary_EmptyShowsNoneAndNote(t *testing.T) {
	inv := analyze.Inventory{
		SchemaVersion: 1,
		Graph:         analyze.DependencyGraph{Kind: "go-packages", Note: "go list failed: x"},
	}
	out := RenderInventorySummary(inv)
	if !strings.Contains(out, "primary language: (none)") {
		t.Fatalf("empty primary must render (none): %s", out)
	}
	if !strings.Contains(out, "(none)") || !strings.Contains(out, "build:") {
		t.Fatalf("empty build/test must render (none): %s", out)
	}
	if !strings.Contains(out, "graph note: go list failed: x") {
		t.Fatalf("graph note must render: %s", out)
	}
}
