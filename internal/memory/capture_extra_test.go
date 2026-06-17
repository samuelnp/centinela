package memory

import (
	"os"
	"testing"
	"time"
)

// sourceFor: plan step maps to decisions parser.
func TestSourceForPlanStep(t *testing.T) {
	spec, ok := sourceFor("alpha", "plan")
	if !ok {
		t.Fatal("expected plan step to have a source")
	}
	if spec.path != "docs/features/alpha.md" {
		t.Fatalf("unexpected plan source path: %q", spec.path)
	}
	if spec.parse == nil {
		t.Fatal("expected non-nil parser for plan step")
	}
}

// sourceFor: docs step has no capture source.
func TestSourceForDocsStep(t *testing.T) {
	_, ok := sourceFor("alpha", "docs")
	if ok {
		t.Fatal("expected docs step to have no capture source")
	}
}

// sourceFor: validate step maps to gatekeeper path.
func TestSourceForValidateStep(t *testing.T) {
	spec, ok := sourceFor("alpha", "validate")
	if !ok {
		t.Fatal("expected validate step to have a source")
	}
	if spec.path != ".workflow/alpha-gatekeeper.md" {
		t.Fatalf("unexpected validate source path: %q", spec.path)
	}
}

// SC-03: Capture for plan step creates decision entries.
func TestCapturePlanStepWritesDecisions(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll("docs/features", 0o755) //nolint:errcheck
	text := "## Decisions\n- use postgres\n- cap at 10\n- no embeddings\n"
	os.WriteFile("docs/features/alpha.md", []byte(text), 0o644) //nolint:errcheck

	Capture("alpha", "plan", enabledCfg())

	entries := loadEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 decision entries from plan step, got %d", len(entries))
	}
}

// SC-07: malformed lesson (no bullets) → warn, no entry written.
func TestCaptureMalformedLessonNoEntry(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0o755) //nolint:errcheck
	// Edge cases file has no bullet points — parseLesson will return error.
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("# Edge Cases\nNo bullets here.\n"), 0o644) //nolint:errcheck

	Capture("alpha", "tests", enabledCfg())

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected no entries for malformed lesson, got %d", len(entries))
	}
}

// persist: duplicate entries produce exactly 1 file (tests idempotence via persist).
func TestPersistIdempotent(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("f", "tests", TypeLesson, "- body", "s", nil, time.Now())
	persist([]Entry{e})
	persist([]Entry{e}) // second call must not duplicate

	entries := loadEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after double persist, got %d", len(entries))
	}
}
