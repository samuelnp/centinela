package memory

import (
	"testing"
	"time"
)

var testTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// SC-01: edge-case lessons parsed into a single lesson entry.
func TestParseLessonValidContent(t *testing.T) {
	text := "# Edge Cases\n- timeout on retry\n- duplicate webhook\n"
	entries, err := parseLesson("alpha", ".workflow/alpha-edge-cases.md", text, testTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 lesson entry, got %d", len(entries))
	}
	if entries[0].Type != TypeLesson {
		t.Fatalf("expected type lesson, got %q", entries[0].Type)
	}
	if entries[0].Feature != "alpha" {
		t.Fatalf("feature mismatch: %q", entries[0].Feature)
	}
}

// SC-07: malformed lesson (no bullets) returns an error (non-blocking in Capture).
func TestParseLessonEmptyBullets(t *testing.T) {
	_, err := parseLesson("alpha", ".workflow/alpha-edge-cases.md", "# No bullets here", testTime)
	if err == nil {
		t.Fatal("expected error for lesson with no bullets")
	}
}

// SC-02: gatekeeper verdict captured into a single verdict entry.
func TestParseVerdictValid(t *testing.T) {
	text := "## Gatekeeper\nAll checks passed. Status: SAFE\n"
	entries, err := parseVerdict("alpha", ".workflow/alpha-gatekeeper.md", text, testTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 verdict entry, got %d", len(entries))
	}
	if entries[0].Type != TypeVerdict {
		t.Fatalf("expected type verdict, got %q", entries[0].Type)
	}
}

// SC-07: empty gatekeeper report returns error.
func TestParseVerdictEmpty(t *testing.T) {
	_, err := parseVerdict("alpha", ".workflow/alpha-gatekeeper.md", "   ", testTime)
	if err == nil {
		t.Fatal("expected error for empty verdict")
	}
}

// SC-03: 3 decision bullets → 3 entries.
func TestParseDecisionsThreeBullets(t *testing.T) {
	text := "## Some section\n\n## Decisions\n- use postgres\n- no embeddings\n- cap at 10 entries\n\n## Other\n"
	entries, err := parseDecisions("alpha", "docs/features/alpha.md", text, testTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 decision entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Type != TypeDecision {
			t.Fatalf("expected decision type, got %q", e.Type)
		}
	}
}

// SC-04: no Decisions section → 0 entries, no error.
func TestParseDecisionsNoSection(t *testing.T) {
	text := "## Problem\nsome problem\n## Goals\n- do this\n"
	entries, err := parseDecisions("alpha", "docs/features/alpha.md", text, testTime)
	if err != nil {
		t.Fatalf("unexpected error for missing Decisions: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
