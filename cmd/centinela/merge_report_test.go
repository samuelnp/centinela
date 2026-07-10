package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// Defense in depth: an outcome carrying neither proof flag must be refused —
// never rendered as the fabricated success line this feature killed.
func TestReportMergeSuccess_NoProof_Refuses(t *testing.T) {
	var err error
	printed := captureStdout(t, func() {
		err = reportMergeSuccess(worktree.MergeOutcome{Feature: "high-score"})
	})
	if err == nil || !strings.Contains(err.Error(), "refusing to claim success") {
		t.Fatalf("unverified outcome must be refused, got: %v", err)
	}
	if strings.Contains(printed, "Merged") || strings.Contains(printed, "already merged") {
		t.Fatalf("no success wording may be printed without proof: %s", printed)
	}
}

// AlreadyMerged gets the honest wording, not the delivery claim.
func TestReportMergeSuccess_AlreadyMerged_HonestWording(t *testing.T) {
	out := captureStdout(t, func() {
		if err := reportMergeSuccess(worktree.MergeOutcome{Feature: "high-score", AlreadyMerged: true}); err != nil {
			t.Fatalf("already-merged outcome must report cleanly: %v", err)
		}
	})
	if !strings.Contains(out, `"high-score" was already merged`) {
		t.Fatalf("want the already-merged wording, got: %s", out)
	}
	if strings.Contains(out, `Merged "high-score" into main`) {
		t.Fatalf("must not fabricate a fresh delivery: %s", out)
	}
}

// A verified ref advance gets the full delivery confirmation, worded with
// the actual target branch (EC-01: a non-"main" primary must not be lied to).
func TestReportMergeSuccess_RefAdvanced_DeliveryWording(t *testing.T) {
	out := captureStdout(t, func() {
		if err := reportMergeSuccess(worktree.MergeOutcome{Feature: "high-score", RefAdvanced: true, TargetBranch: "trunk"}); err != nil {
			t.Fatalf("advanced outcome must report cleanly: %v", err)
		}
	})
	if !strings.Contains(out, `Merged "high-score" into trunk and removed its worktree.`) {
		t.Fatalf("want the delivery confirmation with the real target branch, got: %s", out)
	}
}
