package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookPlanAdvisorOnlyDuringPlan(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	workflow.Save(wf) //nolint:errcheck
	planOut := captureStdout(t, func() { withStdin(t, "{}", func() { runHookPlanAdvisor(nil, nil) }) })
	if !strings.Contains(planOut, "CENTINELA PLAN ADVISOR") {
		t.Fatalf("expected plan advisor output, got: %s", planOut)
	}
	wf.CurrentStep = "code"
	workflow.Save(wf) //nolint:errcheck
	codeOut := captureStdout(t, func() { withStdin(t, "{}", func() { runHookPlanAdvisor(nil, nil) }) })
	if strings.TrimSpace(codeOut) != "" {
		t.Fatalf("expected no advisor output outside plan, got: %s", codeOut)
	}
}
