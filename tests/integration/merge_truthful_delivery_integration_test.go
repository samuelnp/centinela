package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func mtdiPass(string) (bool, string) { return true, "" }

// Full truthful-delivery flow against real git: resolve the primary tree
// from inside the worktree, merge there, and observe the ref + the disk.
// A rerun of the same delivery reports AlreadyMerged without moving main.
func TestMergeTruthfulFlow_AdvanceThenHonestRerun(t *testing.T) {
	repo, wt := mtdiRepo(t, "high-score")
	primary, err := worktree.PrimaryTree(wt)
	if err != nil {
		t.Fatalf("PrimaryTree: %v", err)
	}
	before := mtdiGit(t, repo, "rev-parse", "main")
	out, err := worktree.Merge(primary, "high-score", mtdiPass)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("fresh delivery must advance the ref, got %+v", out)
	}
	if mtdiGit(t, repo, "rev-parse", "main") == before {
		t.Fatal("main did not advance in the primary tree")
	}
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be removed; err=%v", e)
	}

	rerun, err := worktree.Merge(primary, "high-score", mtdiPass)
	if err != nil {
		t.Fatalf("rerun must succeed honestly: %v", err)
	}
	if rerun.RefAdvanced || !rerun.AlreadyMerged {
		t.Fatalf("rerun must report AlreadyMerged only, got %+v", rerun)
	}
}

// A text conflict stops the flow with the worktree kept and neither success
// flag set — the steward path never carries a claimable outcome.
func TestMergeTruthfulFlow_ConflictKeepsWorktreeNoSuccessFlags(t *testing.T) {
	repo, wt := mtdiRepo(t, "clash")
	// Diverge main on the same file the feature touched.
	if err := os.WriteFile(filepath.Join(repo, "feat.txt"), []byte("main version\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mtdiGit(t, repo, "add", ".")
	mtdiGit(t, repo, "commit", "-q", "-m", "main edit")
	out, err := worktree.Merge(repo, "clash", mtdiPass)
	if err != nil {
		t.Fatalf("conflict is an outcome, not an error: %v", err)
	}
	if !out.TextConflict || !out.WorktreeKept {
		t.Fatalf("want conflict outcome with worktree kept, got %+v", out)
	}
	if out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("conflict must not carry success flags: %+v", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("worktree must be kept on conflict: %v", e)
	}
}
