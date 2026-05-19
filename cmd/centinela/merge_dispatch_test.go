package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Regression (patch 1): a stale pending marker from a prior stalled merge
// must be cleared when a later clean merge of the same feature succeeds.
func TestRunMerge_CleanMergeClearsStaleMarker(t *testing.T) {
	d := stewardRepo(t, "gamma", false) // no conflict → clean merge
	chdir(t, d)
	// Simulate a leftover marker from an earlier stalled attempt.
	if err := worktree.WritePending(".",
		worktree.MergeOutcome{Feature: "gamma", TextConflict: true}); err != nil {
		t.Fatalf("seed stale marker: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(".", "gamma")); err != nil {
		t.Fatalf("precondition: stale marker must exist: %v", err)
	}
	if err := runMerge(nil, []string{"gamma"}); err != nil {
		t.Fatalf("clean merge must exit zero: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(".", "gamma")); !os.IsNotExist(err) {
		t.Fatalf("stale marker must be cleared by clean merge; err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(d, ".worktrees", "gamma")); !os.IsNotExist(err) {
		t.Fatalf("worktree must be removed by clean merge; err=%v", err)
	}
}

// dispatchSteward writes the marker, prints the directive (naming prompt,
// feature and resume cmd) and returns a non-zero (error) result.
func TestDispatchSteward_WritesMarkerAndDirective(t *testing.T) {
	d := stewardRepo(t, "delta", true)
	chdir(t, d)
	o := worktree.MergeOutcome{Feature: "delta", TextConflict: true,
		ConflictedPaths: []string{"shared.txt"}}
	err := dispatchSteward(o)
	if err == nil {
		t.Fatal("dispatchSteward must return a non-nil (non-zero) error")
	}
	if !strings.Contains(err.Error(), "centinela merge --continue delta") {
		t.Fatalf("error must name resume cmd: %v", err)
	}
	m, lerr := worktree.LoadPending(".", "delta")
	if lerr != nil || m == nil {
		t.Fatalf("marker must be written: %v / %v", m, lerr)
	}
	if m.Reason != "git-text-conflict" {
		t.Fatalf("marker reason = %q", m.Reason)
	}
}
