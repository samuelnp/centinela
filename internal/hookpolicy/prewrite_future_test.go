package hookpolicy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// A state file written by a NEWER Centinela, whose body this binary cannot
// model, must never block a write. This binary does not understand that file's
// step semantics; enforcing a guess would block every governed write in the
// repo, which is exactly the bricking this feature exists to prevent. The
// fixture sits on "plan", where a code file is normally BLOCKED — so the allow
// can only come from the unmodellable rule.
func TestUnmodellableWorkflowNeverBlocksAWrite(t *testing.T) {
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
	if !d.Allow {
		t.Fatalf("a write governed by an unmodellable workflow must be allowed, got %+v", d)
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
