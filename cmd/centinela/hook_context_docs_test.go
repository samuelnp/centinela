package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// A user-facing docs step nags for the documentation-specialist evidence file
// (markdown-first contract) AND the changelog — never for the deleted portal.
func TestRunHookContextDocsReminder(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "docs"
	workflow.Save(wf) //nolint:errcheck
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "Documentation output missing") {
		t.Fatalf("expected docs reminder, got: %s", out)
	}
	if !strings.Contains(out, "Changelog entry missing") {
		t.Fatalf("changelog nudge now applies to every feature, got: %s", out)
	}
	// The specialist evidence file silences the docs nudge — no portal involved.
	os.WriteFile(".workflow/f-documentation-specialist.md", []byte("# docs"), 0644) //nolint:errcheck
	out = captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
	if strings.Contains(out, "Documentation output missing") {
		t.Fatalf("specialist evidence file must silence the docs nudge: %s", out)
	}
	if !strings.Contains(out, "Changelog entry missing") {
		t.Fatalf("changelog nudge must persist until the changelog exists: %s", out)
	}
}
