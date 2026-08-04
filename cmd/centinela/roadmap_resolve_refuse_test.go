package main

import (
	"os"
	"strings"
	"testing"
)

// AC13: refusal leaves the markers on disk and stages nothing.
func TestResolveCommandRefusalLeavesMarkers(t *testing.T) {
	resolveRepo(t, markers)
	var staged []string
	base := `{"phases":[{"name":"Phase 13: Lighter Centinela","features":[]}]}`
	run := resolveStub(
		map[string]string{".workflow/roadmap.json": "unmerged\n"},
		map[string]string{
			":1:.workflow/roadmap.json": base,
			":2:.workflow/roadmap.json": `{"phases":[{"name":"Phase 13: Lighter Centinela","features":[{"name":"o"}]}]}`,
			":3:.workflow/roadmap.json": `{"phases":[{"name":"Phase 13: Lighter Centinela","features":[{"name":"t"}]}]}`,
		}, &staged)
	err := resolveRoadmapState(".", run)
	if err == nil || !strings.Contains(err.Error(), "Phase 13: Lighter Centinela") {
		t.Fatalf("want a named refusal, got %v", err)
	}
	body, _ := os.ReadFile(".workflow/roadmap.json") //nolint:errcheck
	if string(body) != markers {
		t.Fatalf("roadmap.json must be byte-identical:\n%s", body)
	}
	if len(staged) != 0 {
		t.Fatalf("nothing may be staged: %v", staged)
	}
	if _, err := os.Stat("ROADMAP.md"); err == nil {
		t.Fatal("a refused resolve must not regenerate ROADMAP.md")
	}
}
