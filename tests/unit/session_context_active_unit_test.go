package unit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

// writeWF saves a real <feature>.json workflow-state file at the given step.
func writeWF(t *testing.T, feature, step string) {
	t.Helper()
	wf := workflow.New(feature)
	wf.CurrentStep = step
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save %s: %v", feature, err)
	}
}

// chdirWorkflowTemp creates a temp dir with a .workflow/ subdir and chdirs into it.
func chdirWorkflowTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return d
}

// Spec scenarios 1/2/3: only a genuine non-done <feature>.json survives;
// evidence JSONs (feature != basename, empty step), done workflows, and ad-hoc
// roadmap JSONs are all rejected.
func TestActiveWorkflows_RejectsNoiseKeepsRealNonDone(t *testing.T) {
	chdirWorkflowTemp(t)
	writeWF(t, "alpha", "code") // real, non-done -> kept
	writeWF(t, "beta", "done")  // done -> rejected
	// evidence JSON: feature != basename, no currentStep field.
	ev := filepath.Join(workflow.WorkflowDir, "alpha-qa-senior.json")
	os.WriteFile(ev, []byte(`{"feature":"alpha","role":"qa-senior"}`), 0644) //nolint:errcheck
	// ad-hoc roadmap JSONs whose basename matches no feature field.
	rj := filepath.Join(workflow.WorkflowDir, "roadmap.json")
	os.WriteFile(rj, []byte(`{"phases":[]}`), 0644) //nolint:errcheck
	rq := filepath.Join(workflow.WorkflowDir, "roadmap-quality.json")
	os.WriteFile(rq, []byte(`{"role":"roadmap-quality-evaluator"}`), 0644) //nolint:errcheck

	got := workflow.ActiveWorkflows(workflow.WorkflowDir)
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 active workflow, got %d: %+v", len(got), got)
	}
	if got[0].Feature != "alpha" {
		t.Fatalf("expected surviving feature alpha, got %q", got[0].Feature)
	}
}

// Spec scenario 4: duplicate evidence JSONs for a feature do not multiply the
// row; the single genuine workflow-state file appears exactly once.
func TestActiveWorkflows_DedupesToSingleRow(t *testing.T) {
	chdirWorkflowTemp(t)
	writeWF(t, "epsilon", "code")
	for _, role := range []string{"big-thinker", "qa-senior", "senior-engineer"} {
		p := filepath.Join(workflow.WorkflowDir, "epsilon-"+role+".json")
		os.WriteFile(p, []byte(`{"feature":"epsilon"}`), 0644) //nolint:errcheck
	}
	got := workflow.ActiveWorkflows(workflow.WorkflowDir)
	count := 0
	for _, wf := range got {
		if wf.Feature == "epsilon" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("epsilon should appear exactly once, got %d", count)
	}
}

// Spec scenario 5 (ordering half): survivors are sorted by file mtime
// descending — most-recently-touched first.
func TestActiveWorkflows_SortsByMtimeDescending(t *testing.T) {
	chdirWorkflowTemp(t)
	order := []string{"oldest", "middle", "newest"}
	base := time.Now().Add(-3 * time.Hour)
	for i, f := range order {
		writeWF(t, f, "tests")
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(workflow.FilePath(f), mt, mt) //nolint:errcheck
	}
	got := workflow.ActiveWorkflows(workflow.WorkflowDir)
	if len(got) != 3 {
		t.Fatalf("expected 3 workflows, got %d", len(got))
	}
	want := []string{"newest", "middle", "oldest"}
	for i, w := range want {
		if got[i].Feature != w {
			t.Fatalf("position %d: want %q, got %q", i, w, got[i].Feature)
		}
	}
}
