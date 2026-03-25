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

func TestRunHookAutostartSkipsWhenPromptIsNotFeatureIntent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                   //nolint:errcheck
	os.Chdir(d)                                                         //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing"), 0644) //nolint:errcheck
	withStdin(t, `{"prompt":"step plan is done shall I advance?"}`, func() {
		runHookAutostart(nil, nil) //nolint:errcheck
	})
	if _, err := os.Stat(workflow.WorkflowDir); !os.IsNotExist(err) {
		t.Fatal("expected no workflow directory for review prompt")
	}
}

func TestUniqueFeatureNameAddsSuffixOnCollision(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                   //nolint:errcheck
	os.Chdir(d)                                                         //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                             //nolint:errcheck
	os.WriteFile(workflow.FilePath("release-flow"), []byte("{}"), 0644) //nolint:errcheck
	if got := uniqueFeatureName("release-flow"); got != "release-flow-2" {
		t.Fatalf("expected collision suffix, got %q", got)
	}
}

func TestRunHookAutostartSkipsWhenStartGuardsFail(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	withStdin(t, `{"prompt":"please add release diagnostics"}`, func() {
		runHookAutostart(nil, nil) //nolint:errcheck
	})
	if _, err := os.Stat(workflow.WorkflowDir); !os.IsNotExist(err) {
		t.Fatal("expected no workflow when PROJECT.md guard fails")
	}
}
