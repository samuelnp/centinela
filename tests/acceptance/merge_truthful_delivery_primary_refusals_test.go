package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge refuses when the primary tree is in detached HEAD state
func TestAccMergeDetachedPrimaryHeadRefused(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	mtdaGit(t, repo, "checkout", "-q", "--detach")
	before := mtdaGit(t, repo, "rev-parse", "HEAD")

	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a detached primary tree must be refused (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "detached HEAD") {
		t.Fatalf("the refusal must name the detached HEAD:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "HEAD") != before {
		t.Fatal("advancing a detached HEAD is not \"main advanced\" — it must not move")
	}
	if strings.Contains(out, mtdcSuccess) || strings.Contains(out, "already merged") {
		t.Fatalf("no success message on refusal:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("the worktree must survive the refusal: %v", e)
	}
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: merge refuses when the primary working tree is bare
func TestAccMergeBarePrimaryRefused(t *testing.T) {
	bin := buildCent(t)
	wt := mtdfBarePrimaryRepo(t, "high-score")

	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a bare primary tree must be refused (code=%d):\n%s", code, out)
	}
	if !strings.Contains(out, "primary working tree is bare") {
		t.Fatalf("the refusal must mention the bare primary working tree:\n%s", out)
	}
	if strings.Contains(out, mtdcSuccess) || strings.Contains(out, "already merged") {
		t.Fatalf("no success message on refusal:\n%s", out)
	}
	if _, e := os.Stat(wt); e != nil {
		t.Fatalf("the worktree must survive the refusal: %v", e)
	}
}
