package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Re-running the merge for an already-landed branch (worktree already gone)
// exits cleanly with the honest already-merged wording, prints no worktree
// notice (the CWD is the primary tree), and does not move main.
func TestRunMerge_AlreadyMergedRerun_HonestNoNotice(t *testing.T) {
	d := seedCleanMergeRepo(t, "omega") // cwd = d (primary tree)
	stubPortalRegen(t)
	if err := runMerge(nil, []string{"omega"}); err != nil {
		t.Fatalf("first merge must succeed: %v", err)
	}
	before := gitOutTruthful(t, d, "rev-parse", "main")
	var err error
	out := captureStdout(t, func() { err = runMerge(nil, []string{"omega"}) })
	if err != nil {
		t.Fatalf("already-merged rerun must succeed honestly: %v", err)
	}
	if !strings.Contains(out, `"omega" was already merged`) {
		t.Fatalf("want the already-merged wording, got: %s", out)
	}
	if strings.Contains(out, `Merged "omega" into main`) {
		t.Fatalf("rerun must not fabricate a fresh delivery: %s", out)
	}
	if strings.Contains(out, "Merging in primary working tree") {
		t.Fatalf("no worktree notice when run from the primary tree: %s", out)
	}
	if gitOutTruthful(t, d, "rev-parse", "main") != before {
		t.Fatal("rerun must not move main")
	}
}

// A stale pending marker that cannot be cleared surfaces as an error after
// an otherwise-clean merge instead of being silently swallowed.
func TestRunMerge_ClearPendingFailure_Surfaced(t *testing.T) {
	d := seedCleanMergeRepo(t, "omega")
	stubPortalRegen(t)
	// Ignore .workflow/ so the occupied marker does not trip the dirty guard.
	if err := os.WriteFile(filepath.Join(d, ".gitignore"), []byte(".worktrees/\n.workflow/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", ".gitignore"}, {"commit", "-q", "-m", "ignore workflow"}} {
		c := exec.Command("git", args...)
		c.Dir = d
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// A non-empty DIRECTORY at the marker path makes os.Remove fail.
	marker := filepath.Join(d, ".workflow", "omega-merge-pending.json")
	if err := os.MkdirAll(filepath.Join(marker, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := runMerge(nil, []string{"omega"})
	if err == nil || !strings.Contains(err.Error(), "cannot clear") {
		t.Fatalf("want the ClearPending failure surfaced, got: %v", err)
	}
}
