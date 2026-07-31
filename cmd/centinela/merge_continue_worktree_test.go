package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stewardEvidenceIn writes schema-valid steward evidence into repo's
// .workflow/ — the primary tree, where the stalled merge lives.
func stewardEvidenceIn(t *testing.T, repo, feature, handoffTo string) {
	t.Helper()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	writeStewardEvidence(t, feature, handoffTo)
}

// commitInProgressMerge resolves and commits the conflicted merge sitting in
// the primary tree — what the Merge Steward actually does on APPLY.
func commitInProgressMerge(t *testing.T, repo string) {
	t.Helper()
	for _, args := range [][]string{{"checkout", "--theirs", "."}, {"add", "-A"},
		{"commit", "--no-edit", "-q"}} {
		c := exec.Command("git", args...)
		c.Dir = repo
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// Finding 1 at the cmd tier: a stall started from the worktree CWD writes its
// marker into the PRIMARY tree, refuses to resume while the conflict is live,
// and resumes truthfully once the merge really landed.
func TestRunMerge_ContinueFromWorktreeCwd_VerifiesBeforeClaiming(t *testing.T) {
	d := stewardRepo(t, "iota", true)
	wt := filepath.Join(d, ".worktrees", "iota")
	chdir(t, wt)
	if err := runMerge(nil, []string{"iota"}); err == nil {
		t.Fatal("a text conflict must stall")
	}
	if _, e := os.Stat(filepath.Join(d, ".workflow", "iota-merge-pending.json")); e != nil {
		t.Fatalf("the marker must land in the primary tree, not the worktree: %v", e)
	}
	mergeContinue = true
	defer func() { mergeContinue = false }()
	stewardEvidenceIn(t, d, "iota", "complete")

	before := gitOutTruthful(t, d, "rev-parse", "main")
	out := captureStdout(t, func() {
		if err := runMerge(nil, []string{"iota"}); err == nil {
			t.Fatal("an unresolved conflict must not resume successfully")
		}
	})
	if strings.Contains(out, "and removed its worktree") {
		t.Fatalf("FABRICATED SUCCESS while main is unmoved: %s", out)
	}
	if gitOutTruthful(t, d, "rev-parse", "main") != before {
		t.Fatal("the refusal must not move main")
	}

	commitInProgressMerge(t, d)
	out = captureStdout(t, func() {
		if err := runMerge(nil, []string{"iota"}); err != nil {
			t.Fatalf("a landed merge must resume from the worktree CWD: %v", err)
		}
	})
	if !strings.Contains(out, `Merged "iota" into main and removed its worktree.`) {
		t.Fatalf("want the verified success line, got: %s", out)
	}
	gitOutTruthful(t, d, "merge-base", "--is-ancestor", "iota", "main")
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("the worktree must be gone; err=%v", e)
	}
}
