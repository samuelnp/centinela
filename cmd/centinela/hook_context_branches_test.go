package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookContextPlanNeedsBrief(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	workflow.Save(wf) //nolint:errcheck

	withStdin(t, "{}", func() {
		if err := runHookContext(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
}
