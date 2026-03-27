package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusLine_NoWorkflowShowsStartHint(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(orig, "..", ".."))
	bin := filepath.Join(d, "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}

	cmd := exec.Command(bin, "hook", "statusline")
	cmd.Dir = d
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook statusline failed: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "WF:none") || !strings.Contains(s, "NEXT:start-feature") {
		t.Fatalf("expected no-workflow hint, got: %s", s)
	}
}
