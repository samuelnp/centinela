package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// captureStdout is defined in hook_cmd_test.go (same package).

// TestCompleteDoneEmitsDeliveryDirective: advancing the final step to done in
// worktree mode with an origin remote prints the delivery directive listing
// both options — and performs no delivery (text only).
func TestCompleteDoneEmitsDeliveryDirective(t *testing.T) {
	deliverRepo(t, true) // chdir into a git repo with origin + config
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wf := workflow.New("feat")
	for _, s := range []string{"plan", "code", "tests", "validate"} {
		wf.Steps[s] = workflow.StepState{Status: "done"}
	}
	wf.Steps["docs"] = workflow.StepState{Status: "in-progress"}
	wf.CurrentStep = "docs"
	wf.WorktreePath = ".worktrees/feat"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow/feat-changelog.md", []byte("- feat: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var err error
	out := captureStdout(t, func() { err = runComplete(nil, []string{"feat"}) })
	if err != nil {
		t.Fatalf("docs completion should advance to done: %v", err)
	}
	for _, want := range []string{"CENTINELA DIRECTIVE:", "deliver feat --via pr", "deliver feat --via merge", "do NOT push or merge"} {
		if !strings.Contains(out, want) {
			t.Fatalf("done output missing %q:\n%s", want, out)
		}
	}
}
