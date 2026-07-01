package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

// TestPromoteInPlace_SlugNotFound errors when the slug is absent everywhere.
func TestPromoteInPlace_SlugNotFound(t *testing.T) {
	crudPromoteDir(t, draftBody)
	before := crudBytes(t, RoadmapFile)
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{Slug: "absent", Scores: scores}); err == nil {
		t.Fatal("absent slug must error")
	}
	if !bytes.Equal(before, crudBytes(t, RoadmapFile)) {
		t.Fatal("failed promote must be byte-identical")
	}
}

// TestPromoteInPlace_MissingArtifacts aborts on preflight, writing nothing.
func TestPromoteInPlace_MissingArtifacts(t *testing.T) {
	crudChdir(t, draftBody) // no analysis/quality artifacts written
	before := crudBytes(t, RoadmapFile)
	scores, _ := ParseScores("9,9,9,9,9,9")
	_, err := Promote(RoadmapFile, PromoteRequest{Slug: "new-widget", Scores: scores})
	if err == nil {
		t.Fatal("missing artifacts must abort the finalize")
	}
	if !bytes.Equal(before, crudBytes(t, RoadmapFile)) {
		t.Fatal("aborted finalize must leave roadmap.json byte-identical (draft intact)")
	}
	if !strings.Contains(string(crudBytes(t, RoadmapFile)), `"draft":true`) {
		t.Fatal("draft flag must survive the aborted finalize")
	}
}

// TestPromoteFromBacklog_Rejections covers the Backlog branch guards.
func TestPromoteFromBacklog_Rejections(t *testing.T) {
	crudPromoteDir(t, crudBody) // crudBody has Backlog/legacy-finding
	scores, _ := ParseScores("9,9,9,9,9,9")
	if _, err := Promote(RoadmapFile, PromoteRequest{Slug: "legacy-finding", Scores: scores}); err == nil ||
		!strings.Contains(err.Error(), "--phase is required") {
		t.Fatalf("Backlog promote without --phase must error, got %v", err)
	}
	_, err := Promote(RoadmapFile, PromoteRequest{Slug: "legacy-finding", Phase: "Phase 9", Scores: scores})
	if err == nil || !strings.Contains(err.Error(), "unknown phase") {
		t.Fatalf("Backlog promote into unknown phase must error, got %v", err)
	}
}
