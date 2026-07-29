package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Finding 1 at the unit tier: an APPLY verdict is the steward's CLAIM. Until
// the branch is an ancestor of the target, ResolveMerge must refuse — it used
// to remove nothing, clear the marker and let the CLI print "Merged … and
// removed its worktree" with the ref untouched.
func TestResolveMerge_ApplyWithoutLandedBranch_Refuses(t *testing.T) {
	repo := resolveRepo(t, "iota")
	writeMarker(t, repo, "iota")
	before := gitOut(t, repo, "rev-parse", "main")
	res, err := worktree.ResolveMerge(repo, "iota", okValidator("complete"))
	if err == nil || !contains(err.Error(), "was not completed") {
		t.Fatalf("APPLY without a landed branch must refuse, got: %v", err)
	}
	if res.Finalized || res.Outcome.RefAdvanced || res.Outcome.AlreadyMerged {
		t.Fatalf("no success flag may survive the refusal: %+v", res)
	}
	if gitOut(t, repo, "rev-parse", "main") != before {
		t.Fatal("refusal must not move the target ref")
	}
	if _, e := os.Stat(filepath.Join(repo, ".worktrees", "iota")); e != nil {
		t.Fatalf("worktree must be kept on refusal: %v", e)
	}
	if _, e := os.Stat(worktree.PendingPath(repo, "iota")); e != nil {
		t.Fatalf("marker must survive so the merge stays resumable: %v", e)
	}
}

// A landed branch resolves with the same verified flags the direct merge path
// produces, so the CLI can route both through one success reporter.
func TestResolveMerge_LandedBranch_CarriesVerifiedOutcome(t *testing.T) {
	repo := resolveRepo(t, "iota")
	writeMarker(t, repo, "iota")
	landBranch(t, repo, "iota")
	res, err := worktree.ResolveMerge(repo, "iota", okValidator("complete"))
	if err != nil {
		t.Fatalf("ResolveMerge: %v", err)
	}
	if !res.Finalized || !res.Outcome.RefAdvanced || res.Outcome.RemoveFailed {
		t.Fatalf("want a finalized, advanced, cleanly-removed outcome: %+v", res)
	}
	if res.Outcome.TargetBranch != "main" || res.Outcome.Feature != "iota" {
		t.Fatalf("outcome must name the real target and feature: %+v", res.Outcome)
	}
}

// Half-success on the continue path: the branch landed but the worktree is
// busy. The outcome must still say the merge landed.
func TestResolveMerge_BusyWorktree_ReportsLandedMerge(t *testing.T) {
	repo := resolveRepo(t, "iota")
	writeMarker(t, repo, "iota")
	landBranch(t, repo, "iota")
	scratch := filepath.Join(repo, ".worktrees", "iota", "scratch.txt")
	if err := os.WriteFile(scratch, []byte("busy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := worktree.ResolveMerge(repo, "iota", okValidator("complete"))
	if err == nil {
		t.Fatal("a busy worktree must not be reported as removed")
	}
	if res.Finalized || !res.Outcome.RemoveFailed || !res.Outcome.RefAdvanced {
		t.Fatalf("want an unfinalized outcome that still records the landing: %+v", res)
	}
}
