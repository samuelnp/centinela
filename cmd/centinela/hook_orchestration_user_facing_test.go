package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookOrchestrationIncludesUXRoleForUserFacingCode(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	workflow.Save(wf) //nolint:errcheck
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "ux-ui-specialist") {
		t.Fatalf("expected ux-ui-specialist in directive, got: %s", out)
	}
}
