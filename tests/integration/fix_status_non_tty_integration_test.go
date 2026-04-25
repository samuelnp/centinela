package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestStatusFallsBackWithoutTTY(t *testing.T) {
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
	old, _ := os.Getwd()
	os.Chdir(d)                          //nolint:errcheck
	workflow.Save(workflow.New("alpha")) //nolint:errcheck
	os.Chdir(old)                        //nolint:errcheck
	cmd := exec.Command(bin, "status", "alpha")
	cmd.Dir = d
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status failed without tty: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "Feature") || !strings.Contains(s, "alpha") {
		t.Fatalf("expected rendered status output, got: %s", s)
	}
	if strings.Contains(s, "/dev/tty") {
		t.Fatalf("unexpected tty error in output: %s", s)
	}
}
