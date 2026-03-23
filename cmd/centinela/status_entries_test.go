package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunStatusAllWithEntriesReturns(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	workflow.Save(workflow.New("f"))        //nolint:errcheck
	_ = runStatusAll(nil, nil)
}
