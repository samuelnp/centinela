package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookPrewriteBlockedPaths(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	exitCode := 0
	oldExit := exitPrewrite
	oldEval := evalPrewrite
	defer func() {
		exitPrewrite = oldExit
		evalPrewrite = oldEval
	}()
	exitPrewrite = func(c int) { exitCode = c }
	evalPrewrite = func(string, string, *config.Config, []*workflow.Workflow) hookpolicy.PrewriteDecision {
		return hookpolicy.PrewriteDecision{NeedInit: true, FileType: workflow.TypeCode}
	}

	withStdin(t, `{"tool_input":{"filePath":"`+d+`/internal/a.go"}}`, func() {
		_ = runHookPrewrite(nil, nil)
	})
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for NeedInit, got %d", exitCode)
	}

	exitCode = 0
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "plan"
	workflow.Save(wf) //nolint:errcheck
	evalPrewrite = func(string, string, *config.Config, []*workflow.Workflow) hookpolicy.PrewriteDecision {
		return hookpolicy.PrewriteDecision{FileType: workflow.TypeCode, Step: "plan", Feature: "f"}
	}
	withStdin(t, `{"tool_input":{"filePath":"`+d+`/internal/a.go"}}`, func() {
		_ = runHookPrewrite(nil, nil)
	})
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for blocked step, got %d", exitCode)
	}
}
