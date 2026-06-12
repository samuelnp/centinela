package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

// End-to-end: `centinela start --profile outcome` pins the profile, and the
// prewrite policy then allows a code write during the plan step — the relaxation
// flows from the persisted state, not a test-only shortcut.
func TestStartOutcomeAllowsCodeDuringPlan(t *testing.T) {
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela-ep-int")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "PROJECT.md"), []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	start := exec.Command(bin, "start", "feat", "--profile", "outcome")
	start.Dir = dir
	if out, err := start.CombinedOutput(); err != nil {
		t.Fatalf("start failed: %v\n%s", err, out)
	}

	wf, err := workflow.Load(filepath.Join(dir, ".workflow", "feat.json"))
	if err != nil {
		// Load resolves relative to CWD; read via chdir instead.
		t.Chdir(dir)
		wf, err = workflow.Load("feat")
		if err != nil {
			t.Fatalf("load workflow: %v", err)
		}
	}
	if wf.EnforcementProfile != config.ProfileOutcome || wf.CurrentStep != "plan" {
		t.Fatalf("expected outcome+plan, got %q/%q", wf.EnforcementProfile, wf.CurrentStep)
	}

	d := hookpolicy.EvaluatePrewrite(filepath.Join(dir, "internal", "x.go"), dir, &config.Config{},
		[]*workflow.Workflow{wf})
	if !d.Allow {
		t.Fatalf("outcome workflow must allow code write during plan, got %+v", d)
	}
}
