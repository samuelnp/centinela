package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestReviseLoopEndToEnd drives a validate→code rewind plus the downstream
// invalidation the cmd layer performs, asserting that certification evidence is
// shed while source and test files survive and the revision is persisted.
func TestReviseLoopEndToEnd(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wf := workflow.New("feat")
	ts := "2026-06-30T00:00:00Z"
	for _, s := range []string{"plan", "code", "tests"} {
		wf.Steps[s] = workflow.StepState{Status: "done", CompletedAt: &ts}
	}
	wf.Steps["validate"] = workflow.StepState{Status: "in-progress"}
	wf.CurrentStep = "validate"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}

	write := func(p string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".workflow/feat-qa-senior.json")
	write(".workflow/feat-validation-specialist.json")
	write(".workflow/feat-edge-cases.md")
	src, tst := "internal/feat/service.go", "tests/unit/feat/service_test.go"
	write(src)
	write(tst)

	reopened, err := wf.RewindTo("code", "bug in service")
	if err != nil {
		t.Fatalf("RewindTo: %v", err)
	}
	seen := map[evidence.Role]bool{}
	for _, step := range reopened {
		roles, arts := evidence.InvalidationTargets("feat", step)
		for _, r := range roles {
			if seen[r] {
				continue
			}
			seen[r] = true
			if _, err := evidence.Invalidate("feat", r); err != nil {
				t.Fatal(err)
			}
		}
		for _, a := range arts {
			if _, err := evidence.InvalidateArtifact("feat", a); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}

	for _, gone := range []string{
		".workflow/feat-qa-senior.json",
		".workflow/feat-validation-specialist.json",
		".workflow/feat-edge-cases.md",
	} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Fatalf("%s must be removed", gone)
		}
	}
	for _, keep := range []string{src, tst} {
		if _, err := os.Stat(keep); err != nil {
			t.Fatalf("%s must survive: %v", keep, err)
		}
	}
	got, err := workflow.Load("feat")
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrentStep != "code" || len(got.Revisions) != 1 {
		t.Fatalf("state = %+v", got)
	}
}
