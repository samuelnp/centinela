package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookContextDocsReminder(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "docs"
	workflow.Save(wf) //nolint:errcheck
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "Documentation output missing") {
		t.Fatalf("expected docs reminder, got: %s", out)
	}
}
