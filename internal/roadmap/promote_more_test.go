package roadmap

import (
	"os"
	"strings"
	"testing"
)

// TestPromote_MissingRoadmapFile returns error immediately.
func TestPromote_MissingRoadmapFile(t *testing.T) {
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote("/nonexistent/roadmap.json", PromoteRequest{
		Slug: "x", Phase: "Phase 5", Scores: scores,
	}); err == nil {
		t.Fatal("expected error for missing roadmap.json")
	}
}

// TestLoadBacklogFinding_MissingFile returns error.
func TestLoadBacklogFinding_MissingFile(t *testing.T) {
	if _, err := LoadBacklogFinding("/nonexistent/roadmap.json", "x"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

// TestPromote_ProvidesSummaryOverride uses override when non-empty.
func TestPromote_ProvidesSummaryOverride(t *testing.T) {
	setupPromoteDir(t)
	scores, _ := ParseScores("9,9,9,9,9,9")
	Promote(RoadmapFile, PromoteRequest{ //nolint:errcheck
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
		Summary: "overridden summary",
	})
	data, _ := os.ReadFile(RoadmapQualityFile)
	if !strings.Contains(string(data), "overridden summary") {
		t.Error("summary override must appear in quality file")
	}
}

// TestPromote_DeferredAtPreservedInProvenance checks RFC3339 round-trip.
func TestPromote_DeferredAtPreservedInProvenance(t *testing.T) {
	setupPromoteDir(t)
	scores, _ := ParseScores("9,9,9,9,9,9")
	Promote(RoadmapFile, PromoteRequest{ //nolint:errcheck
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	})
	analysisData, _ := os.ReadFile(RoadmapAnalysisMarkdown)
	if !strings.Contains(string(analysisData), "2026-01-01T00:00:00Z") {
		t.Errorf("deferredAt must appear in provenance bullet: %s", analysisData)
	}
}
