package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookStatuslinePrintsOutput(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	mkdir(t, workflow.WorkflowDir)
	wf := workflow.New("alpha")
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() {
			if err := runHookStatusline(nil, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if !strings.Contains(out, "WF:alpha") {
		t.Fatalf("expected statusline output, got: %s", out)
	}
}
