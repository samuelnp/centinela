package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-steward-auto-dispatch.feature
// Scenario: Clean merge does not dispatch the Steward (regression guard).
func TestAcceptance_CleanMergeNoDispatch(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := mergeRepo(t, "gamma", false)
	out, err := runBin(t, bin, repo, "merge", "gamma")
	if err != nil {
		t.Fatalf("clean merge must exit zero: %v\n%s", err, out)
	}
	if strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("clean merge must not emit a directive:\n%s", out)
	}
	if _, e := os.Stat(filepath.Join(repo, ".workflow", "gamma-merge-pending.json")); !os.IsNotExist(e) {
		t.Fatalf("clean merge must not write a pending marker; err=%v", e)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "gamma")); !os.IsNotExist(e) {
		t.Fatalf("clean merge must remove the worktree; err=%v", e)
	}
}

// Scenario: Text conflict writes the pending marker and dispatches the Steward.
func TestAcceptance_TextConflictDispatches(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := mergeRepo(t, "delta", true)
	out, err := runBin(t, bin, repo, "merge", "delta")
	if err == nil {
		t.Fatalf("text conflict must exit non-zero:\n%s", out)
	}
	if !strings.Contains(out, "CENTINELA DIRECTIVE:") ||
		!strings.Contains(out, "merge-steward-prompt.md") ||
		!strings.Contains(out, "delta") ||
		!strings.Contains(out, "centinela merge --continue delta") {
		t.Fatalf("directive must name prompt, feature and resume cmd:\n%s", out)
	}
	if _, e := os.Stat(filepath.Join(repo, ".workflow", "delta-merge-pending.json")); e != nil {
		t.Fatalf("pending marker must be written on text conflict: %v", e)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "delta")); e != nil {
		t.Fatalf("worktree must be kept on text conflict: %v", e)
	}
}

// Scenario: Continue with APPLY evidence finalizes the merge.
func TestAcceptance_ContinueApplyFinalizes(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := mergeRepo(t, "iota", true)
	if _, err := runBin(t, bin, repo, "merge", "iota"); err == nil {
		t.Fatal("expected conflict dispatch to exit non-zero")
	}
	abortRepoMerge(t, repo)
	writeMergeEvidence(t, repo, "iota", "complete")
	out, err := runBin(t, bin, repo, "merge", "--continue", "iota")
	if err != nil {
		t.Fatalf("APPLY --continue must finalize: %v\n%s", err, out)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "iota")); !os.IsNotExist(e) {
		t.Fatalf("worktree must be removed on APPLY finalize; err=%v", e)
	}
	if _, e := os.Stat(filepath.Join(repo, ".workflow", "iota-merge-pending.json")); !os.IsNotExist(e) {
		t.Fatalf("marker must be cleared on APPLY finalize; err=%v", e)
	}
}

// Scenario: Continue with ESCALATE evidence keeps the merge blocked.
func TestAcceptance_ContinueEscalateBlocks(t *testing.T) {
	work := t.TempDir()
	bin := buildCentinela(t, work)
	repo := mergeRepo(t, "kappa", true)
	if _, err := runBin(t, bin, repo, "merge", "kappa"); err == nil {
		t.Fatal("expected conflict dispatch to exit non-zero")
	}
	abortRepoMerge(t, repo)
	writeMergeEvidence(t, repo, "kappa", "user")
	out, err := runBin(t, bin, repo, "merge", "--continue", "kappa")
	if err == nil {
		t.Fatalf("ESCALATE --continue must exit non-zero:\n%s", out)
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "kappa")); e != nil {
		t.Fatalf("worktree must be kept on escalate: %v", e)
	}
	if _, e := os.Stat(filepath.Join(repo, ".workflow", "kappa-merge-pending.json")); e != nil {
		t.Fatalf("marker must be kept on escalate: %v", e)
	}
}
