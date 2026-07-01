package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func seedReviseWF(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	wf := workflow.New("f")
	ts := "2026-06-30T00:00:00Z"
	for _, s := range []string{"plan", "code", "tests"} {
		wf.Steps[s] = workflow.StepState{Status: "done", CompletedAt: &ts}
	}
	wf.Steps["validate"] = workflow.StepState{Status: "in-progress"}
	wf.CurrentStep = "validate"
	workflow.Save(wf) //nolint:errcheck
	for _, r := range []string{"qa-senior", "validation-specialist", "gatekeeper"} {
		os.WriteFile(".workflow/f-"+r+".json", []byte("x"), 0o644) //nolint:errcheck
		os.WriteFile(".workflow/f-"+r+".md", []byte("x"), 0o644)   //nolint:errcheck
	}
	os.WriteFile(".workflow/f-edge-cases.md", []byte("x"), 0o644) //nolint:errcheck
}

func TestRunReviseHappyPath(t *testing.T) {
	seedReviseWF(t)
	reviseTo, reviseReason = "code", "bug found"
	captureStdout(t, func() {
		if err := runRevise(nil, []string{"f"}); err != nil {
			t.Fatalf("runRevise: %v", err)
		}
	})
	wf, err := workflow.Load("f")
	if err != nil {
		t.Fatal(err)
	}
	if wf.CurrentStep != "code" || len(wf.Revisions) != 1 {
		t.Fatalf("state = %+v", wf)
	}
	if _, err := os.Stat(".workflow/f-validation-specialist.json"); !os.IsNotExist(err) {
		t.Fatal("validate evidence must be invalidated")
	}
	if _, err := os.Stat(".workflow/f-edge-cases.md"); !os.IsNotExist(err) {
		t.Fatal("edge-cases must be invalidated")
	}
}

func TestRunReviseEmptyReasonRejected(t *testing.T) {
	seedReviseWF(t)
	reviseTo, reviseReason = "code", "   "
	if err := runRevise(nil, []string{"f"}); err == nil {
		t.Fatal("whitespace reason must be rejected")
	}
}

func TestRunReviseMissingWorkflowErrors(t *testing.T) {
	t.Chdir(t.TempDir())
	reviseTo, reviseReason = "code", "x"
	if err := runRevise(nil, []string{"missing"}); err == nil {
		t.Fatal("missing workflow must error")
	}
}

func TestRunReviseRewindRejected(t *testing.T) {
	seedReviseWF(t)
	reviseTo, reviseReason = "docs", "forward target"
	if err := runRevise(nil, []string{"f"}); err == nil {
		t.Fatal("forward target must error")
	}
}
