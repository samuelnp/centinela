package workflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func chdirActiveTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func saveStep(t *testing.T, feature, step string) {
	t.Helper()
	wf := New(feature)
	wf.CurrentStep = step
	if err := Save(wf); err != nil {
		t.Fatalf("save %s: %v", feature, err)
	}
}

// ActiveWorkflows keeps the genuine non-done state file, rejecting evidence
// JSONs, ad-hoc roadmap JSONs and done workflows; survivors sort by mtime desc.
func TestActiveWorkflows_Internal(t *testing.T) {
	chdirActiveTemp(t)
	saveStep(t, "alpha", "code")
	saveStep(t, "beta", "done")
	os.WriteFile(filepath.Join(WorkflowDir, "alpha-qa.json"), []byte(`{"feature":"alpha"}`), 0o644) //nolint:errcheck
	os.WriteFile(filepath.Join(WorkflowDir, "roadmap.json"), []byte(`{"phases":[]}`), 0o644)        //nolint:errcheck
	saveStep(t, "gamma", "tests")
	older := time.Now().Add(-time.Hour)
	os.Chtimes(FilePath("alpha"), older, older) //nolint:errcheck

	got := ActiveWorkflows(WorkflowDir)
	if len(got) != 2 {
		t.Fatalf("expected 2 active, got %d: %+v", len(got), got)
	}
	if got[0].Feature != "gamma" || got[1].Feature != "alpha" {
		t.Fatalf("expected gamma (newest) then alpha, got %q,%q", got[0].Feature, got[1].Feature)
	}
}

// CapActive: above-cap returns front N + omitted; at/below and max<=0 → all, more=0.
func TestCapActive_Internal(t *testing.T) {
	mk := func(n int) []*Workflow {
		out := make([]*Workflow, n)
		for i := range out {
			out[i] = New("f")
		}
		return out
	}
	if shown, more := CapActive(mk(7), 5); len(shown) != 5 || more != 2 {
		t.Fatalf("above cap: want (5,2), got (%d,%d)", len(shown), more)
	}
	if shown, more := CapActive(mk(3), 5); len(shown) != 3 || more != 0 {
		t.Fatalf("below cap: want (3,0), got (%d,%d)", len(shown), more)
	}
	if shown, more := CapActive(mk(4), 0); len(shown) != 4 || more != 0 {
		t.Fatalf("no cap: want (4,0), got (%d,%d)", len(shown), more)
	}
}
