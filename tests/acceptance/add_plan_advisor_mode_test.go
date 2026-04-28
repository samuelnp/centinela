package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/add-plan-advisor-mode.feature
func TestPlanAdvisorHookActsOnlyDuringPlan(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(d, "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}
	os.Chdir(d)                                                                //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                         //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n"), 0644) //nolint:errcheck
	workflow.Save(workflow.New("f"))                                           //nolint:errcheck
	planCmd := exec.Command(bin, "hook", "plan-advisor")
	planCmd.Dir = d
	planOut, err := planCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("plan-advisor hook failed: %v\n%s", err, planOut)
	}
	s := string(planOut)
	if !strings.Contains(s, "CENTINELA PLAN ADVISOR") || !strings.Contains(s, "Ask at most 4 questions") {
		t.Fatalf("expected advisor output during plan, got: %s", s)
	}
	wf, _ := workflow.Load("f")
	wf.CurrentStep = "code"
	workflow.Save(wf) //nolint:errcheck
	codeCmd := exec.Command(bin, "hook", "plan-advisor")
	codeCmd.Dir = d
	codeOut, err := codeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("plan-advisor hook failed in code step: %v\n%s", err, codeOut)
	}
	if strings.TrimSpace(string(codeOut)) != "" {
		t.Fatalf("expected no advisor output outside plan, got: %s", codeOut)
	}
}
