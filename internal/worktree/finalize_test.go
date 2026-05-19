package worktree_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestResolveMerge_ApplyCleanFinalizes(t *testing.T) {
	repo := resolveRepo(t, "iota")
	writeMarker(t, repo, "iota")
	res, err := worktree.ResolveMerge(repo, "iota", okValidator("complete"))
	if err != nil {
		t.Fatalf("ResolveMerge: %v", err)
	}
	if !res.Finalized || res.Escalated {
		t.Fatalf("APPLY+clean must finalize: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "iota")); !os.IsNotExist(err) {
		t.Fatalf("worktree must be removed on finalize; err=%v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(repo, "iota")); !os.IsNotExist(err) {
		t.Fatalf("marker must be cleared on finalize; err=%v", err)
	}
}

func TestResolveMerge_EscalateKeepsWorktreeAndMarker(t *testing.T) {
	repo := resolveRepo(t, "kappa")
	writeMarker(t, repo, "kappa")
	res, err := worktree.ResolveMerge(repo, "kappa", okValidator("user"))
	if err != nil {
		t.Fatalf("ESCALATE must not error: %v", err)
	}
	if res.Finalized || !res.Escalated {
		t.Fatalf("ESCALATE must block finalize: %+v", res)
	}
	if res.EscalationNote == "" {
		t.Fatal("escalation note must be surfaced")
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "kappa")); err != nil {
		t.Fatalf("worktree must be kept on escalate: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(repo, "kappa")); err != nil {
		t.Fatalf("marker must be kept on escalate: %v", err)
	}
}

func TestResolveMerge_InvalidEvidenceRefuses(t *testing.T) {
	repo := resolveRepo(t, "mu")
	writeMarker(t, repo, "mu")
	bad := func(string) (string, error) { return "", errors.New("schema invalid") }
	res, err := worktree.ResolveMerge(repo, "mu", bad)
	if err == nil {
		t.Fatal("invalid evidence must refuse with error")
	}
	if res.Finalized {
		t.Fatal("must not finalize on invalid evidence")
	}
	if _, err := os.Stat(worktree.PendingPath(repo, "mu")); err != nil {
		t.Fatalf("marker must survive invalid evidence: %v", err)
	}
}

func TestResolveMerge_NoMarkerClearError(t *testing.T) {
	repo := resolveRepo(t, "xi")
	_, err := worktree.ResolveMerge(repo, "xi", okValidator("complete"))
	if err == nil || !contains(err.Error(), "no pending merge to continue") {
		t.Fatalf("no-marker must report clear error, got: %v", err)
	}
}

func TestResolveMerge_DirtyMainRefusesEvenOnApply(t *testing.T) {
	repo := resolveRepo(t, "nu")
	writeMarker(t, repo, "nu")
	// Dirty the main tree with an uncommitted change.
	_ = os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("x\n"), 0o644)
	res, err := worktree.ResolveMerge(repo, "nu", okValidator("complete"))
	if err == nil || !contains(err.Error(), "dirty") {
		t.Fatalf("dirty main must block even on APPLY, got: %v", err)
	}
	if res.Finalized {
		t.Fatal("must not finalize when main is dirty")
	}
}

func TestResolveMerge_CorruptMarkerReturnsError(t *testing.T) {
	repo := resolveRepo(t, "rho")
	_ = os.MkdirAll(filepath.Join(repo, ".workflow"), 0o755)
	_ = os.WriteFile(worktree.PendingPath(repo, "rho"), []byte("{bad"), 0o644)
	_, err := worktree.ResolveMerge(repo, "rho", okValidator("complete"))
	if err == nil {
		t.Fatal("corrupt marker must surface an error")
	}
}
