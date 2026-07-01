package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// setupDraftPromote chdirs into a temp project holding a single draft feature
// plus empty analysis/quality artifacts, and resets the promote flag globals.
func setupDraftPromote(t *testing.T) {
	t.Helper()
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile,
		`{"phases":[{"name":"Phase 1","features":[{"name":"new-widget","draft":true}]}]}`)
	writeFile(t, roadmap.RoadmapAnalysisFile, `{"role":"senior-product-manager","features":[]}`)
	writeFile(t, roadmap.RoadmapQualityFile, `{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`)
	writeFile(t, roadmap.RoadmapAnalysisMarkdown, "# analysis\n")
	writeFile(t, roadmap.RoadmapQualityMarkdown, "# quality\n")
	t.Cleanup(func() { promotePhase, promoteScores, promoteSummary = "", "", "" })
	promotePhase, promoteSummary = "", ""
}

// TestPromoteScored_DraftInPlace finalizes a draft in place and validates.
func TestPromoteScored_DraftInPlace(t *testing.T) {
	setupDraftPromote(t)
	promoteScores = "9,9,9,9,9,9"
	captureStdout(t, func() {
		if err := promoteScored("new-widget"); err != nil {
			t.Fatalf("promoteScored: %v", err)
		}
	})
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if roadmap.IsDraftFeature(r, "new-widget") {
		t.Fatal("draft flag must be cleared by in-place finalize")
	}
}

// TestPromoteScored_DraftLowScoreByteIdentical refuses below-9 with draft intact.
func TestPromoteScored_DraftLowScoreByteIdentical(t *testing.T) {
	setupDraftPromote(t)
	promoteScores = "9,9,9,9,9,8" // overall < 9
	before, _ := os.ReadFile(roadmap.RoadmapFile)
	if err := promoteScored("new-widget"); err == nil ||
		!strings.Contains(err.Error(), "at least 9") {
		t.Fatalf("below-9 must be refused, got %v", err)
	}
	after, _ := os.ReadFile(roadmap.RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Fatal("refused finalize must leave roadmap.json byte-identical (draft intact)")
	}
}

// TestPromoteResultMessage phrases both branches by whether a phase moved.
func TestPromoteResultMessage(t *testing.T) {
	t.Cleanup(func() { promotePhase = "" })
	promotePhase = ""
	if !strings.Contains(promoteResultMessage("x"), "Finalized draft") {
		t.Fatal("empty phase → in-place finalize wording")
	}
	promotePhase = "Phase 1"
	if !strings.Contains(promoteResultMessage("x"), "Promoted") {
		t.Fatal("set phase → moved-promote wording")
	}
}
