package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookAutostartStartsWhenNoActiveWorkflow(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                   //nolint:errcheck
	os.Chdir(d)                                                         //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing"), 0644) //nolint:errcheck
	withStdin(t, `{"prompt":"please add release diagnostics"}`, func() {
		runHookAutostart(nil, nil) //nolint:errcheck
	})
	entries, _ := os.ReadDir(workflow.WorkflowDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(entries))
	}
}

func TestRunHookAutostartSkipsWhenActiveWorkflowExists(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                   //nolint:errcheck
	os.Chdir(d)                                                         //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                             //nolint:errcheck
	workflow.Save(workflow.New("active"))                               //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing"), 0644) //nolint:errcheck
	withStdin(t, `{"prompt":"please add release diagnostics"}`, func() {
		runHookAutostart(nil, nil) //nolint:errcheck
	})
	entries, _ := os.ReadDir(workflow.WorkflowDir)
	if len(entries) != 1 {
		t.Fatalf("expected only existing workflow, got %d", len(entries))
	}
}
