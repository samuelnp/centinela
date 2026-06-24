package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// deliverRepo chdirs into a fresh git repo (no origin) with a minimal config,
// reverting chdir on cleanup. addOrigin adds a dummy origin remote.
func deliverRepo(t *testing.T, addOrigin bool) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	run("init")
	if addOrigin {
		run("remote", "add", "origin", "https://example.com/x.git")
	}
	if err := os.WriteFile(dir+"/centinela.toml", []byte("[workflow]\ndisable_auto_commit=true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir+"/"+workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func setVia(t *testing.T, v string) {
	t.Helper()
	prev := deliverVia
	deliverVia = v
	t.Cleanup(func() { deliverVia = prev })
}

// TestRunDeliverRejectsBadVia: an unsupported --via value errors without side effects.
func TestRunDeliverRejectsBadVia(t *testing.T) {
	setVia(t, "bogus")
	if err := runDeliver(nil, []string{"feat"}); err == nil || !strings.Contains(err.Error(), "choose --via") {
		t.Fatalf("bad via should error with guidance, got %v", err)
	}
}

// TestRunDeliverPRWithoutOrigin: --via pr in a repo with no origin and no
// worktree is refused.
func TestRunDeliverPRWithoutOrigin(t *testing.T) {
	deliverRepo(t, false)
	if err := workflow.Save(workflow.New("feat")); err != nil {
		t.Fatal(err)
	}
	setVia(t, "pr")
	if err := runDeliver(nil, []string{"feat"}); err == nil || !strings.Contains(err.Error(), "no origin remote") {
		t.Fatalf("pr without origin should be refused, got %v", err)
	}
}

// TestRunDeliverMergeWithoutWorktree: --via merge in single-checkout mode (no
// worktree path) is refused even when an origin exists.
func TestRunDeliverMergeWithoutWorktree(t *testing.T) {
	deliverRepo(t, true)
	if err := workflow.Save(workflow.New("feat")); err != nil {
		t.Fatal(err)
	}
	setVia(t, "merge")
	if err := runDeliver(nil, []string{"feat"}); err == nil || !strings.Contains(err.Error(), "worktree mode required") {
		t.Fatalf("merge without worktree should be refused, got %v", err)
	}
}
