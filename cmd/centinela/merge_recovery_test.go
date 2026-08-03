package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// A failure with nothing verified is passed through untouched — only a merge
// that demonstrably landed may state that it landed.
func TestReportMergeFailure_UnverifiedOutcomePassesThrough(t *testing.T) {
	boom := errors.New("boom")
	if got := reportMergeFailure(worktree.MergeOutcome{Feature: "f"}, "cmd", boom); got != boom {
		t.Fatalf("want the original error, got: %v", got)
	}
	half := worktree.MergeOutcome{Feature: "f", RemoveFailed: true}
	if got := reportMergeFailure(half, "cmd", boom); got != boom {
		t.Fatalf("a removal failure without a verified advance must pass through: %v", got)
	}
	advanced := worktree.MergeOutcome{Feature: "f", RefAdvanced: true}
	if got := reportMergeFailure(advanced, "cmd", boom); got != boom {
		t.Fatalf("a non-removal failure must pass through: %v", got)
	}
}

// The half-success message must state what IS true, name the cause, and name
// the command that finishes the job.
func TestReportMergeFailure_HalfSuccessStatesBothHalves(t *testing.T) {
	o := worktree.MergeOutcome{Feature: "high-score", TargetBranch: "main",
		RefAdvanced: true, RemoveFailed: true}
	err := reportMergeFailure(o, "centinela merge high-score",
		errors.New("contains modified or untracked files"))
	for _, want := range []string{
		`"high-score" merged into main — verified`,
		"worktree removal failed: contains modified or untracked files",
		"re-run `centinela merge high-score --force-remove` to retry removal",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("message missing %q: %v", want, err)
		}
	}
}

// The idempotent re-run says "already merged", not a fresh delivery.
func TestReportMergeFailure_AlreadyMergedWording(t *testing.T) {
	o := worktree.MergeOutcome{Feature: "high-score", TargetBranch: "main",
		AlreadyMerged: true, RemoveFailed: true}
	err := reportMergeFailure(o, "centinela merge high-score", errors.New("busy"))
	if !strings.Contains(err.Error(), "already merged into main — verified") {
		t.Fatalf("want the already-merged wording: %v", err)
	}
	if strings.Contains(err.Error(), `"high-score" merged into main`) {
		t.Fatalf("a re-run must not read as a fresh delivery: %v", err)
	}
}

func TestMergeRemovalOpts_FollowsTheFlag(t *testing.T) {
	if opts := mergeRemovalOpts(); opts != nil {
		t.Fatalf("force removal must be opt-in, got %d option(s)", len(opts))
	}
	mergeForceRemove = true
	defer func() { mergeForceRemove = false }()
	if opts := mergeRemovalOpts(); len(opts) != 1 {
		t.Fatalf("--force-remove must pass the option through, got %d", len(opts))
	}
}

func TestInDir_UnreachableDirectoryIsAnError(t *testing.T) {
	err := inDir(filepath.Join(t.TempDir(), "absent"), func() error { return nil })
	if err == nil || !strings.Contains(err.Error(), "cannot enter") {
		t.Fatalf("an unreachable directory must error clearly, got: %v", err)
	}
}
