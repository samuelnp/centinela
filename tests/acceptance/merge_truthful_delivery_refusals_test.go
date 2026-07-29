package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge refuses when the primary tree cannot be resolved
func TestAccMergeRefusesOutsideAnyRepo(t *testing.T) {
	bin := buildCent(t)
	out, code := runCent(t, bin, t.TempDir(), "merge", "high-score")
	if code == 0 {
		t.Fatalf("merge outside a repo must exit non-zero:\n%s", out)
	}
	if !strings.Contains(out, "cannot resolve primary working tree") {
		t.Fatalf("want the never-guess refusal:\n%s", out)
	}
	if strings.Contains(out, `Merged "high-score" into main`) {
		t.Fatalf("no success message on refusal:\n%s", out)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: dirty primary tree is still refused
func TestAccMergeDirtyPrimaryRefusedFromWorktreeCwd(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	writeFile(t, repo, "shared.txt", "uncommitted edit\n") // tracked file modified
	before := mtdaGit(t, repo, "rev-parse", "main")
	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code == 0 {
		t.Fatalf("dirty primary must refuse (the guard now checks the RIGHT tree):\n%s", out)
	}
	if !strings.Contains(out, "dirty") {
		t.Fatalf("refusal must say the main working tree is dirty:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") != before {
		t.Fatal("main must be unchanged on refusal")
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("worktree must still exist on refusal: %v", e)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge refuses when the primary tree has the feature branch checked out
func TestAccMergeSelfMergeRefusedNotAlreadyMerged(t *testing.T) {
	bin := buildCent(t)
	repo := mtdaSelfMergeRepo(t, "high-score")
	out, code := runCent(t, bin, repo, "merge", "high-score")
	if code == 0 {
		t.Fatalf("self-merge must exit non-zero:\n%s", out)
	}
	if strings.Contains(out, "already merged") {
		t.Fatalf("self-merge must NOT be reported as already merged:\n%s", out)
	}
	if strings.Contains(out, `Merged "high-score" into main`) {
		t.Fatalf("no success message on self-merge refusal:\n%s", out)
	}
	if !strings.Contains(out, "cannot merge a branch into itself") {
		t.Fatalf("want the self-merge refusal wording:\n%s", out)
	}
}
