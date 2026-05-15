package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func setupMergeRepo(t *testing.T, feature string) (repo, wt string) {
	t.Helper()
	repo = initRepoForWorktrees(t)
	var err error
	wt, err = worktree.Create(repo, feature)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	path := filepath.Join(wt, "feat-"+feature+".txt")
	if err := os.WriteFile(path, []byte("from "+feature), 0644); err != nil {
		t.Fatalf("write feature file: %v", err)
	}
	gitRun(t, wt, "add", ".")
	gitRun(t, wt, "commit", "-q", "-m", "feature commit "+feature)
	return repo, wt
}

func TestMerge_CleanRemovesWorktree(t *testing.T) {
	repo, wt := setupMergeRepo(t, "gamma")
	passing := func(_ string) (bool, string) { return true, "ok" }
	out, err := worktree.Merge(repo, "gamma", passing)
	if err != nil {
		t.Fatalf("Merge clean: %v", err)
	}
	if out.TextConflict || out.ValidateFail {
		t.Fatalf("unexpected flags on clean merge: %+v", out)
	}
	if out.WorktreeKept {
		t.Fatal("WorktreeKept should be false on clean success")
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree dir should be removed; err=%v", err)
	}
}

func TestMerge_DirtyMain_FailsFast(t *testing.T) {
	repo, _ := setupMergeRepo(t, "kappa")
	if err := os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("oops"), 0644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	gitRun(t, repo, "add", "dirty.txt")
	called := false
	runner := func(_ string) (bool, string) { called = true; return true, "" }
	_, err := worktree.Merge(repo, "kappa", runner)
	if err == nil {
		t.Fatal("Merge must fail fast on dirty main")
	}
	if called {
		t.Fatal("validator must not run when pre-check fails")
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "kappa")); err != nil {
		t.Fatalf("worktree should be untouched after dirty pre-check: %v", err)
	}
}
