package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: Writing a feature brief without an active workflow is allowed
func TestRoadmapPhase_FeatureBriefIsRoadmapType(t *testing.T) {
	cfg := &config.Config{}
	got := workflow.ClassifyFile("/project/docs/features/caesar-cipher.md", cfg)
	if got != workflow.TypeRoadmap {
		t.Errorf("feature brief should be TypeRoadmap so hook allows it, got %q", got)
	}
}

// Scenario: Writing ROADMAP.md without an active workflow is allowed
func TestRoadmapPhase_RoadmapMdIsRoadmapType(t *testing.T) {
	cfg := &config.Config{}
	got := workflow.ClassifyFile("/project/ROADMAP.md", cfg)
	if got != workflow.TypeRoadmap {
		t.Errorf("ROADMAP.md should be TypeRoadmap, got %q", got)
	}
}

// Scenario: Feature brief writes are still allowed during plan step
func TestRoadmapPhase_FeatureBriefAllowedInPlanStep(t *testing.T) {
	if !workflow.IsAllowedInStep(workflow.TypeRoadmap, "plan") {
		t.Error("TypeRoadmap must be allowed in plan step")
	}
}
