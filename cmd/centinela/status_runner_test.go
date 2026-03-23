package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunStatusAndStatusAllWithRunner(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	workflow.Save(workflow.New("f"))        //nolint:errcheck

	old := statusRunner
	defer func() { statusRunner = old }()
	called := 0
	statusRunner = func(_ []*workflow.Workflow) error { called++; return nil }
	if err := runStatus(nil, []string{"f"}); err != nil {
		t.Fatalf("runStatus error: %v", err)
	}
	if err := runStatusAll(nil, nil); err != nil {
		t.Fatalf("runStatusAll error: %v", err)
	}
	if called < 2 {
		t.Fatalf("expected status runner called twice, got %d", called)
	}
}
