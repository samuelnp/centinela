package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: removal is verified against git's worktree registry, not a path convention
func TestAccMergeWorktreeOutsideConventionIsReallyRemoved(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	elsewhere := filepath.Join(t.TempDir(), "relocated")
	mtdaGit(t, repo, "worktree", "move", filepath.Join(repo, ".worktrees", "high-score"), elsewhere)

	out, code := runCent(t, bin, repo, "merge", "high-score")
	if strings.Contains(out, mtdcSuccess) {
		if _, e := os.Stat(elsewhere); !os.IsNotExist(e) {
			t.Fatalf("FALSE CLAIM — worktree still on disk at %s:\n%s", elsewhere, out)
		}
		if reg := mtdaGit(t, repo, "worktree", "list", "--porcelain"); strings.Contains(reg, elsewhere) {
			t.Fatalf("FALSE CLAIM — worktree still registered:\n%s", reg)
		}
	}
	if code != 0 {
		t.Fatalf("a relocated worktree must still be delivered (code=%d):\n%s", code, out)
	}
	mtdaGit(t, repo, "merge-base", "--is-ancestor", "high-score", "main")
}

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: the post-merge validate runs against the merged primary tree
func TestAccMergeValidatesMergedPrimaryTreeNotInvokingCwd(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code != 0 {
		t.Fatalf("merge from the worktree CWD must exit 0 (code=%d):\n%s", code, out)
	}
	if strings.Contains(out, "0 files changed") {
		t.Fatalf("the merge-time gate must not self-neuter on an empty diff:\n%s", out)
	}
	if !strings.Contains(out, "Built-in Gates (full scan)") {
		t.Fatalf("the merged tree must be scanned in full, not diff-aware:\n%s", out)
	}
}

// Acceptance: specs/docs-step-markdown-first.feature
// Scenario: The HTML pipeline is gone
// (Supersedes the merge-truthful-delivery portal-regeneration scenario: no
// code path regenerates docs/project-docs/index.html at merge time.)
func TestAccMergeDoesNotRegeneratePortalInPrimaryTree(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	writeFile(t, repo, "PROJECT.md", "# Project\n")
	writeFile(t, repo, "ROADMAP.md", "# Roadmap\n")
	mtdaGit(t, repo, "add", "-A")
	mtdaGit(t, repo, "commit", "-q", "-m", "doc inputs")
	writeFile(t, repo, ".workflow/roadmap.json",
		`{"phases":[{"name":"Phase 1","features":[{"slug":"high-score","status":"done"}]}]}`)

	out, code := runCent(t, bin, wt, "merge", "high-score")
	if code != 0 {
		t.Fatalf("merge must exit 0 (code=%d):\n%s", code, out)
	}
	if _, e := os.Stat(filepath.Join(repo, "docs", "project-docs", "index.html")); !os.IsNotExist(e) {
		t.Fatalf("no portal may be regenerated at merge time (stat err=%v):\n%s", e, out)
	}
	if strings.Contains(out, "portal regen") {
		t.Fatalf("merge output must not mention portal regeneration:\n%s", out)
	}
}
