package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestHookContext_RemindsEdgeCaseReportInTestsStep(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(orig, "..", ".."))
	bin := filepath.Join(d, "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}

	os.MkdirAll(filepath.Join(d, workflow.WorkflowDir), 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "tests"
	oldDir, _ := os.Getwd()
	os.Chdir(d)       //nolint:errcheck
	workflow.Save(wf) //nolint:errcheck
	os.Chdir(oldDir)  //nolint:errcheck

	cmd := exec.Command(bin, "hook", "context")
	cmd.Dir = d
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook context execution failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Edge-case report missing") {
		t.Fatalf("expected edge-case reminder in output: %s", out)
	}
}
