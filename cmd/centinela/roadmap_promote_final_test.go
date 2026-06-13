package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestPromoteScored_PromoteError returns error when Promote() fails (no Backlog entry).
func TestPromoteScored_PromoteError(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = "Phase 5"
	promoteScores = "9,9,9,9,9,9"
	// Use a slug that is NOT in the Backlog phase -> Promote returns "not a Backlog finding"
	if err := promoteScored("does-not-exist-in-backlog"); err == nil {
		t.Fatal("expected error when slug not in Backlog")
	}
}

// TestReportPromoteResult_QualityValidateFails errors when quality validate fails.
func TestReportPromoteResult_QualityValidateFails(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)           //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	// Roadmap with "slug" in a phase
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 5", Features: []roadmap.Feature{{Name: "slug"}}},
	}}
	roadmap.Save(r) //nolint:errcheck
	// Analysis passes (covers "slug")
	writeRoadmapAnalysis(t, "slug")
	// Quality file is present but with WRONG role -> ValidateQuality fails
	os.WriteFile(roadmap.RoadmapQualityMarkdown, []byte("# q\n"), 0644)                                                                                                                                                                                              //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"wrong-role","threshold":9,"features":[{"name":"slug","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}]}`), 0644) //nolint:errcheck
	err := reportPromoteResult("slug")
	if err == nil {
		t.Fatal("expected error when quality role wrong")
	}
	if !strings.Contains(err.Error(), "validate") {
		t.Errorf("error should mention validate, got: %v", err)
	}
}
