package verdict

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func writeEvidence(t *testing.T, feature, role, body string) {
	t.Helper()
	p := filepath.Join(workflow.WorkflowDir, feature+"-"+role+".json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdirTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

// EvidenceIndex lists one entry per on-disk role file, carries the contract
// fields, and is sorted by role name regardless of write order.
func TestEvidenceIndex_SortedByRole(t *testing.T) {
	chdirTemp(t)
	writeEvidence(t, "feat", "qa-senior",
		`{"feature":"feat","step":"tests","role":"qa-senior","status":"done","generatedAt":"2026-06-12T00:00:00Z","handoffTo":"validation-specialist"}`)
	writeEvidence(t, "feat", "big-thinker",
		`{"feature":"feat","step":"plan","role":"big-thinker","status":"done","generatedAt":"2026-06-11T00:00:00Z","handoffTo":"feature-specialist"}`)

	idx := EvidenceIndex("feat")
	if len(idx) != 2 {
		t.Fatalf("want 2 entries, got %d", len(idx))
	}
	if idx[0].Role != "big-thinker" || idx[1].Role != "qa-senior" {
		t.Fatalf("not sorted by role: %v / %v", idx[0].Role, idx[1].Role)
	}
	e := idx[0]
	if e.Step != "plan" || e.Status != "done" || e.HandoffTo != "feature-specialist" {
		t.Fatalf("fields wrong: %+v", e)
	}
	if e.GeneratedAt != "2026-06-11T00:00:00Z" || e.Path != ".workflow/feat-big-thinker.json" {
		t.Fatalf("generatedAt/path wrong: %+v", e)
	}
}

// With no on-disk role files the index is an empty (non-nil) slice so the
// packet still marshals a JSON array.
func TestEvidenceIndex_EmptyNonNil(t *testing.T) {
	chdirTemp(t)
	idx := EvidenceIndex("feat")
	if idx == nil {
		t.Fatal("EvidenceIndex must return a non-nil slice")
	}
	if len(idx) != 0 {
		t.Fatalf("want 0 entries, got %d", len(idx))
	}
}
