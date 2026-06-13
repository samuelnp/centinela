package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookOrchestrationIncludesDocsRole(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	// User-facing feature: the docs step still requires the documentation-specialist.
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "docs"
	workflow.Save(wf) //nolint:errcheck
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "documentation-specialist") {
		t.Fatalf("expected docs role directive, got: %s", out)
	}
}
