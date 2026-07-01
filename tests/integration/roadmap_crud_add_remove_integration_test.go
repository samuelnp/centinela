package integration_test

// Acceptance: specs/roadmap-crud-add-remove.feature
// Scenario: promote finalizes a draft in place — no phase move, draft cleared, artifacts written
// Scenario: remove is refused when the only dependent is itself a draft

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// intoProject chdirs into a temp project seeded with body plus empty artifacts.
func intoProject(t *testing.T, body string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(p, c string) {
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(roadmap.RoadmapFile, body)
	write(roadmap.RoadmapAnalysisFile, `{"role":"senior-product-manager","features":[]}`)
	write(roadmap.RoadmapQualityFile, `{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`)
	write(roadmap.RoadmapAnalysisMarkdown, "# a\n")
	write(roadmap.RoadmapQualityMarkdown, "# q\n")
}

// TestAddThenPromoteInPlaceValidates authors a draft, finalizes it in place, and
// confirms analysis/quality validation then passes — crossing add→promote→validate.
func TestAddThenPromoteInPlaceValidates(t *testing.T) {
	intoProject(t, `{"phases":[{"name":"Phase 1","features":[]}]}`)
	if err := roadmap.Add(roadmap.RoadmapFile, roadmap.AddRequest{Slug: "widget", Phase: "Phase 1"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	scores, _ := roadmap.ParseScores("9,9,9,9,9,9")
	if _, err := roadmap.Promote(roadmap.RoadmapFile, roadmap.PromoteRequest{Slug: "widget", Scores: scores}); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if roadmap.IsDraftFeature(r, "widget") {
		t.Fatal("draft must be cleared after in-place finalize")
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		t.Fatalf("ValidateAnalysis: %v", err)
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		t.Fatalf("ValidateQuality: %v", err)
	}
}

// TestRemoveRefusedByDraftDependent confirms a draft is a real dependent.
func TestRemoveRefusedByDraftDependent(t *testing.T) {
	intoProject(t, `{"phases":[{"name":"Phase 1","features":[`+
		`{"name":"auth-service"},`+
		`{"name":"draft-consumer","dependsOn":["auth-service"],"draft":true}]}]}`)
	err := roadmap.Remove(roadmap.RoadmapFile, "auth-service")
	if err == nil || !strings.Contains(err.Error(), "draft-consumer") {
		t.Fatalf("a draft dependent must block remove, got %v", err)
	}
}
