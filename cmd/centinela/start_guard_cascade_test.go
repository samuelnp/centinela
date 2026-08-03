package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// coldStart lays down the MINIMUM guided greenfield tree: PROJECT.md plus a
// roadmap json with a bootstrap phase. Deliberately no ROADMAP.md, no roadmap
// analysis and no quality report.
func coldStart(t *testing.T, phases ...roadmap.Phase) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	roadmap.Save(&roadmap.Roadmap{Phases: phases})                          //nolint:errcheck
}

func bootstrapPhase(features ...roadmap.Feature) roadmap.Phase {
	return roadmap.Phase{Name: "Phase 0: Bootstrap", Features: features}
}

// TestGuidedColdStartNeedsOnlyProjectAndRoadmap is AC6: the whole point of the
// slimmed cascade.
func TestGuidedColdStartNeedsOnlyProjectAndRoadmap(t *testing.T) {
	coldStart(t, bootstrapPhase(roadmap.Feature{Name: "setup"}))
	order, err := workflowOrderForFeature("setup", config.ProfileGuided)
	if err != nil {
		t.Fatalf("guided cold start must succeed on PROJECT.md + roadmap.json, got %v", err)
	}
	if len(order) != 4 {
		t.Fatalf("expected the bootstrap step order, got %v", order)
	}
}

// TestStrictColdStartStillDemandsTheCascade is AC7 — the ❌ direction, and the
// regression guard that the flip did not silently relax strict.
func TestStrictColdStartStillDemandsTheCascade(t *testing.T) {
	coldStart(t, bootstrapPhase(roadmap.Feature{Name: "setup"}))
	_, err := workflowOrderForFeature("setup", config.ProfileStrict)
	if err == nil || !strings.Contains(err.Error(), "roadmap senior PM analysis") {
		t.Fatalf("strict must still refuse, naming the roadmap analysis; got %v", err)
	}
}
