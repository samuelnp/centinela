package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunCompleteDoneAndValidatePath(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[workflow]\ndisable_auto_commit=true\n[validate]\ncommands=[\"true\"]\n"), 0644) //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                                                 //nolint:errcheck
	os.WriteFile(".workflow/f-gatekeeper.md", []byte("SAFE"), 0644)                                                         //nolint:errcheck

	wf := workflow.New("f")
	wf.CurrentStep = "done"
	workflow.Save(wf) //nolint:errcheck
	if err := runComplete(nil, []string{"f"}); err != nil {
		t.Fatalf("done path should pass: %v", err)
	}

	wf2 := workflow.New("f2")
	wf2.CurrentStep = "validate"
	wf2.Steps["plan"] = workflow.StepState{Status: "done"}
	wf2.Steps["code"] = workflow.StepState{Status: "done"}
	wf2.Steps["tests"] = workflow.StepState{Status: "done"}
	wf2.Steps["validate"] = workflow.StepState{Status: "in-progress"}
	workflow.Save(wf2)                                               //nolint:errcheck
	os.WriteFile(".workflow/f2-gatekeeper.md", []byte("SAFE"), 0644) //nolint:errcheck
	if err := runComplete(nil, []string{"f2"}); err != nil {
		t.Fatalf("validate completion should pass: %v", err)
	}
}

func TestRunCompleteMissingWorkflow(t *testing.T) {
	if err := runComplete(nil, []string{"missing-workflow"}); err == nil {
		t.Fatal("expected missing workflow error")
	}
}
