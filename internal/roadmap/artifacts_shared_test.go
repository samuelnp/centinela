package roadmap

import (
	"strings"
	"testing"
)

// TestAppendScoreArtifacts_WritesBoth records the feature in analysis + quality.
func TestAppendScoreArtifacts_WritesBoth(t *testing.T) {
	crudPromoteDir(t, draftBody)
	scores, _ := ParseScores("9,9,9,9,9,9")
	if err := appendScoreArtifacts("widget", "sum", scores, "- bullet"); err != nil {
		t.Fatalf("appendScoreArtifacts: %v", err)
	}
	if !strings.Contains(string(crudBytes(t, RoadmapAnalysisFile)), "widget") {
		t.Fatal("analysis must gain the entry")
	}
	q := string(crudBytes(t, RoadmapQualityFile))
	if !strings.Contains(q, "widget") || !strings.Contains(q, "sum") {
		t.Fatalf("quality must gain the entry with summary: %s", q)
	}
	if !strings.Contains(string(crudBytes(t, RoadmapAnalysisMarkdown)), "- bullet") {
		t.Fatal("analysis markdown must gain the bullet")
	}
}

// TestAppendScoreArtifacts_MissingFileErrors surfaces the write failure.
func TestAppendScoreArtifacts_MissingFileErrors(t *testing.T) {
	crudChdir(t, draftBody) // no artifact files present
	scores, _ := ParseScores("9,9,9,9,9,9")
	if err := appendScoreArtifacts("widget", "sum", scores, "- b"); err == nil {
		t.Fatal("appendScoreArtifacts must error when artifact files are absent")
	}
}
