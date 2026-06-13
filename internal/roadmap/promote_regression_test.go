package roadmap

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestPromote_PreflightMissingAnalysisJSON leaves roadmap.json unchanged (regression).
func TestPromote_PreflightMissingAnalysisJSON(t *testing.T) {
	setupPromoteDir(t)
	before, _ := os.ReadFile(RoadmapFile)
	os.Remove(RoadmapAnalysisFile) //nolint:errcheck
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	}); err == nil {
		t.Fatal("expected error when analysis.json missing")
	}
	after, _ := os.ReadFile(RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be byte-identical when analysis.json missing (partial-write regression)")
	}
}

// TestPromote_PreflightMissingQualityJSON leaves roadmap.json unchanged (regression).
func TestPromote_PreflightMissingQualityJSON(t *testing.T) {
	setupPromoteDir(t)
	before, _ := os.ReadFile(RoadmapFile)
	os.Remove(RoadmapQualityFile) //nolint:errcheck
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	}); err == nil {
		t.Fatal("expected error when quality.json missing")
	}
	after, _ := os.ReadFile(RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be byte-identical when quality.json missing (partial-write regression)")
	}
}

// TestPromote_UnknownPhase returns error without writes (regression: zero writes).
func TestPromote_UnknownPhase(t *testing.T) {
	setupPromoteDir(t)
	before, _ := os.ReadFile(RoadmapFile)
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 99", Scores: scores,
	}); err == nil {
		t.Fatal("expected error for unknown phase")
	}
	after, _ := os.ReadFile(RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on unknown phase")
	}
}

// TestPromote_SlugNotInBacklog returns error cleanly.
func TestPromote_SlugNotInBacklog(t *testing.T) {
	setupPromoteDir(t)
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "not-in-backlog", Phase: "Phase 5", Scores: scores,
	}); err == nil {
		t.Fatal("expected error for slug not in Backlog")
	}
}

// TestPromote_EmptiedBacklogKept verifies Backlog phase remains after promote.
func TestPromote_EmptiedBacklogKept(t *testing.T) {
	setupPromoteDir(t)
	scores, _ := ParseScores("9,9,9,9,9,9")
	Promote(RoadmapFile, PromoteRequest{Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores}) //nolint:errcheck
	data, _ := os.ReadFile(RoadmapFile)
	if !strings.Contains(string(data), "Backlog") {
		t.Error("empty Backlog phase must be kept in roadmap.json")
	}
}
