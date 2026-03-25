package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookOrchestrationOutputsDirectiveForStrictWorkflow(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "plan"
	workflow.Save(wf) //nolint:errcheck
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "orchestrator only") || !strings.Contains(out, "big-thinker") {
		t.Fatalf("expected strict orchestration directive, got: %s", out)
	}
}
