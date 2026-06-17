package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// An internal docs step renders the changelog banner, not the documentation one.
func TestRunHookContextDocsInternalChangelogReminder(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	os.MkdirAll("docs/features", 0755)      //nolint:errcheck
	// Internal feature: no surface line.
	os.WriteFile("docs/features/f.md", []byte("# f\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "docs"
	workflow.Save(wf) //nolint:errcheck

	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
	if !strings.Contains(out, "Changelog entry missing") {
		t.Fatalf("expected changelog reminder for internal feature, got: %s", out)
	}
	if strings.Contains(out, "Documentation output missing") {
		t.Fatalf("internal feature must not nag for the portal: %s", out)
	}
}

// With the changelog present the internal docs banner stays silent.
func TestRunHookContextDocsInternalSilentWhenChangelogPresent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                //nolint:errcheck
	os.Chdir(d)                                                                      //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                          //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                               //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# f\n"), 0644)                        //nolint:errcheck
	os.WriteFile(workflow.WorkflowDir+"/f-changelog.md", []byte("- fix: x\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "docs"
	workflow.Save(wf) //nolint:errcheck

	out := captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
	if strings.Contains(out, "Changelog entry missing") {
		t.Fatalf("changelog present must silence the banner, got: %s", out)
	}
}
