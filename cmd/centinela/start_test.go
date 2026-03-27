package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunStartValidations(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if err := runStart(nil, []string{"f"}); err == nil {
		t.Fatal("expected PROJECT.md error")
	}
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)       //nolint:errcheck
	workflow.Save(workflow.New("f"))              //nolint:errcheck
	if err := runStart(nil, []string{"f"}); err == nil {
		t.Fatal("expected duplicate workflow error")
	}
}

func TestRunStartCreatesWorkflow(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	if err := runStart(nil, []string{"f2"}); err != nil {
		t.Fatalf("runStart error: %v", err)
	}
	if _, err := os.Stat(workflow.FilePath("f2")); err != nil {
		t.Fatalf("workflow not created: %v", err)
	}
}

func TestStepArrow(t *testing.T) {
	if stepArrow([]string{"plan", "code", "validate", "docs"}) != "plan → code → validate → docs" {
		t.Fatal("expected three-step arrow")
	}
	if stepArrow([]string{"plan", "code", "tests", "validate", "docs"}) != "plan → code → tests → validate → docs" {
		t.Fatal("expected four-step arrow")
	}
}
