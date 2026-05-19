package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestRunMergeContinue_NoMarkerCleanError(t *testing.T) {
	d := stewardRepo(t, "xi", false)
	chdir(t, d)
	err := runMergeContinue("xi")
	if err == nil || !strings.Contains(err.Error(), "no pending merge to continue") {
		t.Fatalf("no-marker continue must error clearly, got: %v", err)
	}
}

func TestRunMergeContinue_ApplyFinalizes(t *testing.T) {
	d := stewardRepo(t, "iota", true)
	chdir(t, d)
	if err := dispatchSteward(worktree.MergeOutcome{
		Feature: "iota", TextConflict: true}); err == nil {
		t.Fatal("dispatch should report block")
	}
	writeStewardEvidence(t, "iota", "complete")
	if err := runMergeContinue("iota"); err != nil {
		t.Fatalf("APPLY continue must finalize: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, ".worktrees", "iota")); !os.IsNotExist(err) {
		t.Fatalf("worktree must be removed: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(".", "iota")); !os.IsNotExist(err) {
		t.Fatalf("marker must be cleared: %v", err)
	}
}

func TestRunMergeContinue_EscalateBlocks(t *testing.T) {
	d := stewardRepo(t, "kappa", true)
	chdir(t, d)
	if err := dispatchSteward(worktree.MergeOutcome{
		Feature: "kappa", TextConflict: true}); err == nil {
		t.Fatal("dispatch should report block")
	}
	writeStewardEvidence(t, "kappa", "user")
	err := runMergeContinue("kappa")
	if err == nil || !strings.Contains(err.Error(), "escalated") {
		t.Fatalf("ESCALATE continue must stay blocked, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, ".worktrees", "kappa")); err != nil {
		t.Fatalf("worktree must be kept on escalate: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(".", "kappa")); err != nil {
		t.Fatalf("marker must be kept on escalate: %v", err)
	}
}

func TestRunMergeContinue_MissingEvidenceRefuses(t *testing.T) {
	d := stewardRepo(t, "lambda", true)
	chdir(t, d)
	if err := dispatchSteward(worktree.MergeOutcome{
		Feature: "lambda", TextConflict: true}); err == nil {
		t.Fatal("dispatch should report block")
	}
	// No steward evidence written.
	err := runMergeContinue("lambda")
	if err == nil || !strings.Contains(err.Error(), "steward evidence required") {
		t.Fatalf("missing evidence must refuse with actionable error, got: %v", err)
	}
}
