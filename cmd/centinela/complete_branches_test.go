package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunCompleteValidateErrorAndWarningBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "validate"
	wf.Steps["plan"] = workflow.StepState{Status: "done"}
	wf.Steps["code"] = workflow.StepState{Status: "done"}
	wf.Steps["tests"] = workflow.StepState{Status: "done"}
	wf.Steps["validate"] = workflow.StepState{Status: "in-progress"}
	workflow.Save(wf)                                                                                                                                                                                                                                                                                 //nolint:errcheck
	os.WriteFile(".workflow/f-gatekeeper.md", []byte("SAFE"), 0644)                                                                                                                                                                                                                                   //nolint:errcheck
	os.MkdirAll("docs/project-docs", 0755)                                                                                                                                                                                                                                                            //nolint:errcheck
	os.WriteFile("docs/project-docs/index.html", []byte("html"), 0644)                                                                                                                                                                                                                                //nolint:errcheck
	os.WriteFile(".workflow/f-validation-specialist.md", []byte("# validation"), 0644)                                                                                                                                                                                                                //nolint:errcheck
	os.WriteFile(".workflow/f-validation-specialist.json", []byte(`{"feature":"f","step":"validate","role":"validation-specialist","status":"done","generatedAt":"2026-05-09T00:00:00Z","inputs":["i"],"outputs":["docs/project-docs/index.html"],"edgeCases":[],"handoffTo":"orchestrator"}`), 0644) //nolint:errcheck
	os.WriteFile("centinela.toml", []byte("[validate]\ncommands=[\"false\"]\n"), 0644)                                                                                                                                                                                                                //nolint:errcheck
	if err := runComplete(nil, []string{"f"}); err == nil {
		t.Fatal("expected validate command failure")
	}
	os.WriteFile("centinela.toml", []byte("[gates]\nproduction_readiness=true\n[validate]\ncommands=[\"true\"]\n"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f-production-readiness.md", []byte("**Status:** WARNING"), 0644)                              //nolint:errcheck
	out := captureStdout(t, func() {
		if err := runComplete(nil, []string{"f"}); err != nil {
			t.Fatalf("expected validate completion success: %v", err)
		}
	})
	if !strings.Contains(out, "WARNING") {
		t.Fatalf("expected production readiness warning output, got: %s", out)
	}
}
