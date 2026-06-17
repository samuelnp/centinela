package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRunRoadmapPromote_ScoredPath_Success promotes and reports results.
func TestRunRoadmapPromote_ScoredPath_Success(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = "Phase 5"
	promoteScores = "9,9,9,9,9,9"
	promoteSummary = ""
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	cmd.Flags().String("phase", "", "")
	cmd.Flags().String("summary", "", "")
	cmd.Flags().Set("scores", promoteScores) //nolint:errcheck
	if err := runRoadmapPromote(cmd, []string{"my-finding"}); err != nil {
		t.Fatalf("scored promote must succeed: %v", err)
	}
	// Check that slug moved to Phase 5
	data, _ := os.ReadFile(roadmap.RoadmapFile)
	if !strings.Contains(string(data), "Phase 5") {
		t.Error("Phase 5 must be present in roadmap")
	}
}

// TestReportPromoteResult_Success verifies validate passes after promotion.
func TestReportPromoteResult_Success(t *testing.T) {
	setupPromoteCmd(t)
	// Promote first to set up the state
	scores, _ := roadmap.ParseScores("9,9,9,9,9,9")
	roadmap.Promote(roadmap.RoadmapFile, roadmap.PromoteRequest{ //nolint:errcheck
		Slug: "my-finding", Phase: "Phase 5", Scores: scores,
	})
	// Seed a valid analysis and quality covering "my-finding"
	os.WriteFile(roadmap.RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[{"name":"my-finding"}]}`), 0644)                                                                                                                                                       //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"my-finding","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}]}`), 0644) //nolint:errcheck
	if err := reportPromoteResult("my-finding"); err != nil {
		t.Fatalf("reportPromoteResult: %v", err)
	}
}

// TestPrintEvaluatorContext_NotInBacklog returns error.
func TestPrintEvaluatorContext_NotInBacklog(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = "Phase 5"
	if err := printEvaluatorContext("nonexistent"); err == nil {
		t.Fatal("must error when slug not in Backlog")
	}
}
