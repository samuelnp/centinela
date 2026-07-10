package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// gitOut runs git in dir and returns trimmed stdout, failing the test on error.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func passingValidate(string) (bool, string) { return true, "" }

// THE regression (2048-rust retrospective WS1.1): resolving the primary tree
// from INSIDE the feature worktree and merging there must actually advance
// main and actually remove the worktree — asserted on the ref and the disk.
func TestMerge_PrimaryResolvedFromWorktreeCwd_AdvancesMainAndRemovesWorktree(t *testing.T) {
	repo, wt := setupMergeRepo(t, "high-score")
	primary, err := worktree.PrimaryTree(wt) // resolve from inside the worktree
	if err != nil {
		t.Fatalf("PrimaryTree from worktree CWD: %v", err)
	}
	wantRepo, _ := filepath.EvalSymlinks(repo)
	gotRepo, _ := filepath.EvalSymlinks(primary)
	if gotRepo != wantRepo {
		t.Fatalf("PrimaryTree = %q, want the primary checkout %q", gotRepo, wantRepo)
	}
	before := gitOut(t, repo, "rev-parse", "main")
	out, err := worktree.Merge(primary, "high-score", passingValidate)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("want RefAdvanced on a fresh merge, got %+v", out)
	}
	if after := gitOut(t, repo, "rev-parse", "main"); after == before {
		t.Fatal("main did NOT advance in the primary tree — the false-success bug")
	}
	gitRun(t, repo, "merge-base", "--is-ancestor", "high-score", "main")
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree must be gone after delivery; stat err=%v", err)
	}
}

// The inverse guard simulates the OLD bug: Merge pointed at the worktree
// directory itself. It must refuse the self-merge instead of fabricating an
// AlreadyMerged success (a commit is trivially its own ancestor).
func TestMerge_RepoPointedAtWorktree_RefusesSelfMerge(t *testing.T) {
	repo, wt := setupMergeRepo(t, "high-score")
	before := gitOut(t, repo, "rev-parse", "main")
	out, err := worktree.Merge(wt, "high-score", passingValidate)
	if err == nil || !strings.Contains(err.Error(), "cannot merge a branch into itself") {
		t.Fatalf("want self-merge refusal, got: %v", err)
	}
	if out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("the old bug shape must not claim success: %+v", out)
	}
	if gitOut(t, repo, "rev-parse", "main") != before {
		t.Fatal("main must be untouched by the refusal")
	}
	if _, err := os.Stat(wt); err != nil {
		t.Fatalf("worktree must survive the refusal: %v", err)
	}
}
