package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRunRoadmapPromote_ScoredPath_OutOfRange still refuses, with no writes —
// deleting the minimum must not have deleted the 1-10 range.
func TestRunRoadmapPromote_ScoredPath_OutOfRange(t *testing.T) {
	setupPromoteCmd(t)
	before, _ := os.ReadFile(roadmap.RoadmapFile)
	promotePhase = "Phase 5"
	promoteScores = "9,9,8,7,9,11"
	cmd := &cobra.Command{}
	cmd.Flags().String("scores", "", "")
	cmd.Flags().Set("scores", promoteScores) //nolint:errcheck
	if err := runRoadmapPromote(cmd, []string{"my-finding"}); err == nil {
		t.Fatal("an out-of-range score must still error")
	}
	after, _ := os.ReadFile(roadmap.RoadmapFile)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on score rejection")
	}
}
