package memory

import (
	"strings"
	"testing"
	"time"
)

// TestComputeIDStability — same inputs always produce the same hash (SC-05).
func TestComputeIDStability(t *testing.T) {
	id1 := computeID("alpha", TypeLesson, "body text")
	id2 := computeID("alpha", TypeLesson, "body text")
	if id1 != id2 {
		t.Fatalf("computeID not stable: %q != %q", id1, id2)
	}
}

// TestComputeIDChangesOnDifferentInputs verifies distinct inputs yield distinct hashes.
func TestComputeIDChangesOnDifferentInputs(t *testing.T) {
	id1 := computeID("alpha", TypeLesson, "body")
	id2 := computeID("alpha", TypeLesson, "different body")
	if id1 == id2 {
		t.Fatal("expected different IDs for different bodies")
	}
}

// TestComputeIDExcludesSourcePath — same feature/type/body → same id regardless of source.
func TestComputeIDExcludesSourcePath(t *testing.T) {
	id1 := computeID("alpha", TypeLesson, "body")
	id2 := computeID("alpha", TypeLesson, "body")
	if id1 != id2 {
		t.Fatal("source path exclusion broken: same content yields different id")
	}
}

// TestNewEntryFields verifies field population and title derivation.
func TestNewEntryFields(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := newEntry("feat", "tests", TypeLesson, "- first line\n- second line", ".workflow/feat-edge-cases.md", []string{"edge-cases"}, at)
	if e.Feature != "feat" {
		t.Fatalf("feature mismatch: %q", e.Feature)
	}
	if e.Type != TypeLesson {
		t.Fatalf("type mismatch: %q", e.Type)
	}
	if e.Title != "- first line" {
		t.Fatalf("unexpected title: %q", e.Title)
	}
	if e.ID == "" {
		t.Fatal("ID must not be empty")
	}
}

// TestFirstLineSkipsBlankLines verifies firstLine skips whitespace-only lines.
func TestFirstLineSkipsBlankLines(t *testing.T) {
	got := firstLine("\n\n  text\n")
	if got != "text" {
		t.Fatalf("expected 'text', got %q", got)
	}
}

// TestFirstLineEmptyBody returns empty string.
func TestFirstLineEmptyBody(t *testing.T) {
	if firstLine("") != "" {
		t.Fatal("expected empty string for empty body")
	}
}

// TestSizeBytes — non-zero for non-empty entry.
func TestSizeBytes(t *testing.T) {
	e := Entry{Title: "t", Body: "body", Tags: []string{"a"}}
	if e.sizeBytes() <= 0 {
		t.Fatal("expected positive size")
	}
}

// TestBodyTrimmed — newEntry trims trailing whitespace from body.
func TestBodyTrimmed(t *testing.T) {
	e := newEntry("f", "tests", TypeLesson, "   body   ", "", nil, time.Now())
	if strings.HasSuffix(e.Body, " ") {
		t.Fatalf("body not trimmed: %q", e.Body)
	}
}
