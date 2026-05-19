package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// runMerge with the --continue flag set must route into runMergeContinue.
func TestRunMerge_ContinueFlagRoutesToContinue(t *testing.T) {
	d := stewardRepo(t, "pi", false)
	chdir(t, d)
	mergeContinue = true
	defer func() { mergeContinue = false }()
	err := runMerge(nil, []string{"pi"})
	if err == nil || !strings.Contains(err.Error(), "no pending merge to continue") {
		t.Fatalf("--continue with no marker must error via continue path, got: %v", err)
	}
}

// dispatchSteward must surface the WritePending error when the marker
// cannot be written (.workflow occupied by a regular file).
func TestDispatchSteward_WritePendingErrorSurfaced(t *testing.T) {
	d := t.TempDir()
	chdir(t, d)
	_ = os.WriteFile(filepath.Join(d, ".workflow"), []byte("x"), 0o644)
	err := dispatchSteward(worktree.MergeOutcome{
		Feature: "rho", TextConflict: true})
	if err == nil {
		t.Fatal("dispatchSteward must surface WritePending failure")
	}
	if strings.Contains(err.Error(), "Merge Steward review") {
		t.Fatalf("expected the write error, not the block hint: %v", err)
	}
}
