package main

import (
	"os"
	"strings"
	"testing"
)

// TestGuidedPromoteReportsIncompleteGradingAsAdvice: under guided the seeded
// artifacts cover only what was promoted, so the coverage complaint is advice —
// refusing here would fail a promotion that already succeeded on disk.
func TestGuidedPromoteReportsIncompleteGradingAsAdvice(t *testing.T) {
	coldPromoteTree(t, "")
	promotePhase, promoteScores, promoteSummary = "Phase 1", "9,9,9,9,9,9", ""
	out := captureStdout(t, func() {
		if err := promoteScored("finding"); err != nil {
			t.Fatalf("guided promote must not refuse on grading coverage: %v", err)
		}
	})
	if !strings.Contains(out, "Advisory") {
		t.Fatalf("incomplete grading must surface as advice: %q", out)
	}
}

// TestGuidedPromoteSilentWhenGradingIsComplete: the advisory is not noise — a
// project whose artifacts do cover the roadmap gets no line at all.
func TestGuidedPromoteSilentWhenGradingIsComplete(t *testing.T) {
	coldPromoteTree(t, "")
	// Cover the roadmap as it stands BEFORE the promote; promote appends the
	// finding itself, so the artifacts end up complete.
	writeRoadmapAnalysis(t, "setup")
	writeRoadmapQuality(t, 9, "setup")
	promotePhase, promoteScores, promoteSummary = "Phase 1", "9,9,9,9,9,9", ""
	out := captureStdout(t, func() { _ = promoteScored("finding") })
	if strings.Contains(out, "Advisory") {
		t.Fatalf("complete grading must draw no advisory: %q", out)
	}
}

// TestGuidedPromoteSurfacesSeedFailure: a seed that cannot be written refuses
// the promote instead of continuing into a half-written artifact set.
func TestGuidedPromoteSurfacesSeedFailure(t *testing.T) {
	coldPromoteTree(t, "")
	os.RemoveAll(".workflow")                            //nolint:errcheck
	os.WriteFile(".workflow", []byte("not a dir"), 0644) //nolint:errcheck
	promotePhase, promoteScores, promoteSummary = "Phase 1", "9,9,9,9,9,9", ""
	if err := promoteScored("finding"); err == nil {
		t.Fatal("an unwritable artifact path must refuse the promote")
	}
}

// TestStrictPromoteSucceedsWithCompleteGrading is the strict happy path: with
// the artifacts present and covering the roadmap, the post-write validation
// passes and promote reports success — the check is real, not a blanket refusal.
func TestStrictPromoteSucceedsWithCompleteGrading(t *testing.T) {
	coldPromoteTree(t, "[workflow]\nenforcement_profile = \"strict\"\n")
	writeRoadmapAnalysis(t, "setup")
	writeRoadmapQuality(t, 9, "setup")
	promotePhase, promoteScores, promoteSummary = "Phase 1", "9,9,9,9,9,9", ""
	out := captureStdout(t, func() {
		if err := promoteScored("finding"); err != nil {
			t.Fatalf("strict promote must succeed with complete grading: %v", err)
		}
	})
	if !strings.Contains(out, "Promoted") {
		t.Fatalf("expected a success message, got %q", out)
	}
}
