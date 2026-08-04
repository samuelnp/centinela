package hookpolicy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// A state file written by a NEWER Centinela, whose body this binary cannot
// model, REFUSES the write and says why. Passing instead would be a
// self-service bypass: `.workflow/*.json` is an ungoverned write target, so any
// agent could write a future version over its own state file and open the gate.
// The refusal names the feature and the remedy (upgrade the binary) — and
// crucially NeedInit stays false, so the hook never claims no workflow exists
// and hook_autostart never forks a duplicate.
func TestUnmodellableWorkflowRefusesAndNamesTheRemedy(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"schemaVersion":99,"feature":"delta","currentStep":"plan","steps":[]}`
	if err := os.WriteFile(workflow.FilePath("delta"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load("delta")
	if err != nil {
		t.Fatalf("a future-version file must load: %v", err)
	}
	if !wf.Unmodellable() {
		t.Fatal("precondition: the fixture must be unmodellable")
	}

	wfs := []*workflow.Workflow{wf}
	d := EvaluatePrewrite(filepath.Join(dir, "internal", "a.go"), dir, &config.Config{}, wfs)
	if d.Allow {
		t.Fatalf("an unreadable state file must not open the gate, got %+v", d)
	}
	if !d.StaleBinary {
		t.Fatalf("the refusal must say the binary is too old, got %+v", d)
	}
	if d.Feature != "delta" {
		t.Fatalf("the refusal must name the feature, got %q", d.Feature)
	}
	if d.NeedInit {
		t.Fatal("the hook must not claim no workflow has been started")
	}
}

// A future-version file this binary CAN model keeps normal step gating: it is
// understood, so it is still governed.
func TestModellableFutureWorkflowKeepsStepGating(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"schemaVersion":99,"feature":"delta","currentStep":"plan","steps":{},"newKey":1}`
	if err := os.WriteFile(workflow.FilePath("delta"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load("delta")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Unmodellable() {
		t.Fatal("a purely additive future file must still be modellable")
	}
	d := EvaluatePrewrite(filepath.Join(dir, "internal", "a.go"), dir, &config.Config{}, []*workflow.Workflow{wf})
	if d.Allow || d.Step != "plan" {
		t.Fatalf("step gating must still apply to a modellable file, got %+v", d)
	}
}
