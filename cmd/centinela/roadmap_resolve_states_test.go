package main

import (
	"os"
	"strings"
	"testing"
)

// Only ROADMAP.md conflicted: regenerate from the already-merged roadmap.json.
func TestResolveCommandMarkdownOnlyConflict(t *testing.T) {
	resolveRepo(t, doc(`{"name":"f","deferredAt":"2026-01-01T00:00:00Z"}`))
	if err := os.WriteFile("ROADMAP.md", []byte(markers), 0o644); err != nil {
		t.Fatal(err)
	}
	var staged []string
	run := resolveStub(map[string]string{"ROADMAP.md": "unmerged\n"}, nil, &staged)
	if err := resolveRoadmapState(".", run); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile("ROADMAP.md") //nolint:errcheck
	if strings.Contains(string(body), "<<<<") || !strings.Contains(string(body), "# Roadmap") {
		t.Fatalf("ROADMAP.md must be regenerated with no markers:\n%s", body)
	}
	if strings.Join(staged, ",") != ".workflow/roadmap.json,ROADMAP.md" {
		t.Fatalf("staged = %v", staged)
	}
}

// E13: outside a conflict it is an exit-0 no-op that modifies nothing.
func TestResolveCommandIsANoOpOutsideAConflict(t *testing.T) {
	body := doc(`{"name":"f","deferredAt":"2026-01-01T00:00:00Z"}`)
	resolveRepo(t, body)
	var staged []string
	out := captureStdout(t, func() {
		if err := resolveRoadmapState(".", resolveStub(nil, nil, &staged)); err != nil {
			t.Errorf("no-op must exit 0: %v", err)
		}
	})
	if !strings.Contains(out, "Nothing to resolve") {
		t.Fatalf("output = %q", out)
	}
	after, _ := os.ReadFile(".workflow/roadmap.json") //nolint:errcheck
	if string(after) != body {
		t.Fatal("roadmap.json must be byte-identical")
	}
	if _, err := os.Stat("ROADMAP.md"); err == nil {
		t.Fatal("a no-op must not write ROADMAP.md")
	}
	if len(staged) != 0 {
		t.Fatalf("nothing may be staged: %v", staged)
	}
}
