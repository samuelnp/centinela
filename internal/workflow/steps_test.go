package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestStepNumberAndStepIndex(t *testing.T) {
	if StepNumber("plan") != 1 || StepNumber("validate") != 4 || StepNumber("docs") != 5 {
		t.Fatal("unexpected step numbers")
	}
	if stepIndex("none") != -1 {
		t.Fatal("expected -1 for unknown step")
	}
}

func TestCompleteTransitionsToDone(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("docs/plans", 0755)                                             //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                          //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                  //nolint:errcheck
	os.MkdirAll("tests/unit", 0755)                                             //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                                       //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# p"), 0644)                        //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# b"), 0644)                     //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644)                 //nolint:errcheck
	os.WriteFile("tests/unit/x_test.go", []byte("package unit"), 0644)          //nolint:errcheck
	os.WriteFile("tests/acceptance/x_test.go", []byte("package acc"), 0644)     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                              //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644)               //nolint:errcheck
	os.WriteFile(".workflow/f-gatekeeper.md", []byte("SAFE"), 0644)             //nolint:errcheck
	os.MkdirAll("docs/project-docs", 0755)                                      //nolint:errcheck
	os.WriteFile("docs/project-docs/index.html", []byte("<html></html>"), 0644) //nolint:errcheck

	wf := New("f")
	cfg := &config.Config{Workflow: config.WorkflowConfig{DisableAutoCommit: true}, Gates: config.GatesConfig{FileSizeEnabled: false}}
	for i := 0; i < 5; i++ {
		if err := wf.Complete(cfg); err != nil {
			t.Fatalf("Complete #%d error: %v", i, err)
		}
	}
	if wf.CurrentStep != "done" {
		t.Fatalf("expected done, got %s", wf.CurrentStep)
	}
}
