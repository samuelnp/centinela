package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Under headless the plan-advisor hook short-circuits before loading workflows
// and emits nothing, even on a plan-step workflow that would otherwise speak.
func TestRunHookPlanAdvisor_HeadlessSilent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                          //nolint:errcheck
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	wf := workflow.New("f")
	workflow.Save(wf) //nolint:errcheck

	t.Setenv("CENTINELA_HEADLESS", "1")
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() {
			if err := runHookPlanAdvisor(nil, nil); err != nil {
				t.Fatalf("runHookPlanAdvisor: %v", err)
			}
		})
	})
	if strings.TrimSpace(out) != "" {
		t.Fatalf("headless plan advisor must emit nothing, got: %s", out)
	}
}
