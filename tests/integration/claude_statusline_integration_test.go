package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestHookStatusLineShowsStepAndBlocker(t *testing.T) {
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
	wf := workflow.New("alpha")
	wf.CurrentStep = "tests"
	old, _ := os.Getwd()
	os.Chdir(d)       //nolint:errcheck
	workflow.Save(wf) //nolint:errcheck
	os.Chdir(old)     //nolint:errcheck

	cmd := exec.Command(bin, "hook", "statusline")
	cmd.Dir = d
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook statusline failed: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "WF:alpha") || !strings.Contains(s, "STEP:tests") {
		t.Fatalf("expected workflow tokens, got: %s", s)
	}
	if !strings.Contains(s, "BLOCK:MISSING_EDGE_CASES") {
		t.Fatalf("expected missing edge-cases blocker, got: %s", s)
	}
}
