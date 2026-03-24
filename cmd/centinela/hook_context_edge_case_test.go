package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookContext_RemindsMissingEdgeCaseReport(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "tests"
	wf.Steps["plan"] = workflow.StepState{Status: "done"}
	wf.Steps["code"] = workflow.StepState{Status: "done"}
	wf.Steps["tests"] = workflow.StepState{Status: "in-progress"}
	workflow.Save(wf) //nolint:errcheck

	output := captureStdout(t, func() {
		withStdin(t, "{}", func() {
			runHookContext(nil, nil) //nolint:errcheck
		})
	})
	if !strings.Contains(output, "Edge-case report missing") {
		t.Fatalf("expected edge-case reminder, got: %s", output)
	}
}
