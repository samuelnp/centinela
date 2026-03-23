package main

import (
	"errors"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunCompleteConfigAndArtifactErrors(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[bad"), 0644) //nolint:errcheck
	if err := runComplete(nil, []string{"f"}); err == nil {
		t.Fatal("expected config parse error")
	}
	os.WriteFile("centinela.toml", []byte("[workflow]\ndisable_auto_commit=true\n"), 0644) //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                //nolint:errcheck
	workflow.Save(workflow.New("f"))                                                       //nolint:errcheck
	if err := runComplete(nil, []string{"f"}); err == nil {
		t.Fatal("expected artifact validation error")
	}
}

func TestRunCompleteSaveErrorViaSeam(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[workflow]\ndisable_auto_commit=true\n"), 0644) //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                        //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                     //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                             //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)                                     //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("x"), 0644)                                  //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644)                            //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                //nolint:errcheck
	workflow.Save(workflow.New("f"))                                                       //nolint:errcheck

	old := saveWorkflow
	defer func() { saveWorkflow = old }()
	saveWorkflow = func(*workflow.Workflow) error { return errors.New("boom") }
	if err := runComplete(nil, []string{"f"}); err == nil {
		t.Fatal("expected save workflow error")
	}
}
