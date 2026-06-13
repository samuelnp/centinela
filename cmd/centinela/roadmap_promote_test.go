package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func setupPromoteCmd(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	rm := `{"phases":[{"name":"Phase 5","features":[]},{"name":"Backlog","features":[{"name":"my-finding","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}]}]}`
	os.WriteFile(roadmap.RoadmapFile, []byte(rm), 0644)                                                                        //nolint:errcheck
	os.WriteFile(roadmap.RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(roadmap.RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0644)                                                //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityMarkdown, []byte("# quality\n"), 0644)                                                  //nolint:errcheck
}

// TestRunRoadmapPromote_NoScoresEvaluatorPath prints context, no writes.
func TestRunRoadmapPromote_NoScoresEvaluatorPath(t *testing.T) {
	setupPromoteCmd(t)
	before, _ := os.ReadFile(roadmap.RoadmapFile)
	promotePhase = "Phase 5"
	promoteScores = ""
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	// scores flag NOT changed: evaluator path
	if err := runRoadmapPromote(cmd, []string{"my-finding"}); err != nil {
		t.Fatalf("evaluator path must not error: %v", err)
	}
	after, _ := os.ReadFile(roadmap.RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on evaluator path")
	}
}

// TestRunRoadmapPromote_ExplicitEmptyScoresIsError rejects --scores="" (regression).
func TestRunRoadmapPromote_ExplicitEmptyScoresIsError(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = "Phase 5"
	promoteScores = ""
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	cmd.Flags().Set("scores", "") //nolint:errcheck // mark as Changed
	err := runRoadmapPromote(cmd, []string{"my-finding"})
	if err == nil {
		t.Fatal("--scores '' must be a usage error (regression: explicit empty --scores)")
	}
	if !strings.Contains(err.Error(), "--scores") {
		t.Errorf("error must mention --scores, got: %v", err)
	}
}

// TestRunRoadmapPromote_NoPhase returns error immediately.
func TestRunRoadmapPromote_NoPhase(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = ""
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	if err := runRoadmapPromote(cmd, []string{"my-finding"}); err == nil {
		t.Fatal("missing --phase must error")
	}
}

// TestRunRoadmapPromote_ScoredPath_LowScore returns error, no writes.
func TestRunRoadmapPromote_ScoredPath_LowScore(t *testing.T) {
	setupPromoteCmd(t)
	before, _ := os.ReadFile(roadmap.RoadmapFile)
	promotePhase = "Phase 5"
	promoteScores = "9,9,8,7,9,7" // overall=7
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	cmd.Flags().Set("scores", promoteScores) //nolint:errcheck
	if err := runRoadmapPromote(cmd, []string{"my-finding"}); err == nil {
		t.Fatal("low overall score must error")
	}
	after, _ := os.ReadFile(roadmap.RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on score rejection")
	}
}
