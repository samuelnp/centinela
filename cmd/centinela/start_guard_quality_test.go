package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// greenfieldWithGrading lays a greenfield tree whose quality report scores the
// bootstrap feature `overall`.
func greenfieldWithGrading(t *testing.T, overall int) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	roadmap.Save(&roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap",
		Features: []roadmap.Feature{{Name: "setup"}}}}}) //nolint:errcheck
	writeRoadmapAnalysis(t, "setup")
	writeRoadmapQuality(t, overall, "setup")
}

// TestStrictAcceptsLowQualityScore: the threshold deletion is UNCONDITIONAL.
// Even strict — which still demands the artifacts exist — no longer reads a
// minimum out of them.
func TestStrictAcceptsLowQualityScore(t *testing.T) {
	greenfieldWithGrading(t, 3)
	if _, err := workflowOrderForFeature("setup", config.ProfileStrict); err != nil {
		t.Fatalf("strict must accept a low self-assigned score, got %v", err)
	}
}

// TestStrictStillRequiresTheQualityArtifact: what strict still enforces is the
// EXISTENCE of the evaluation, which is a process requirement, not a score bar.
func TestStrictStillRequiresTheQualityArtifact(t *testing.T) {
	greenfieldWithGrading(t, 9)
	os.Remove(".workflow/roadmap-quality.json") //nolint:errcheck
	_, err := workflowOrderForFeature("setup", config.ProfileStrict)
	if err == nil || !strings.Contains(err.Error(), "quality evaluation") {
		t.Fatalf("expected roadmap quality error, got %v", err)
	}
}
