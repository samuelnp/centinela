package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestGuidedStillRefusesWhatTheRoadmapSays is AC8: every refusal derived from
// the roadmap's own content survives the slimming, unchanged.
func TestGuidedStillRefusesWhatTheRoadmapSays(t *testing.T) {
	cases := []struct {
		name    string
		phases  []roadmap.Phase
		feature string
		wantSub string
	}{
		{"backlog finding", []roadmap.Phase{bootstrapPhase(roadmap.Feature{Name: "setup"}),
			{Name: "Backlog", Features: []roadmap.Feature{{Name: "finding"}}}}, "finding", "Backlog finding"},
		{"no bootstrap phase", []roadmap.Phase{{Name: "Phase 1", Features: []roadmap.Feature{{Name: "x"}}}},
			"x", "Phase 0: Bootstrap"},
		{"non-bootstrap while bootstrap incomplete", []roadmap.Phase{
			bootstrapPhase(roadmap.Feature{Name: "setup"}),
			{Name: "Phase 1", Features: []roadmap.Feature{{Name: "later"}}}}, "later", "bootstrap is incomplete"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			coldStart(t, tc.phases...)
			_, err := workflowOrderForFeature(tc.feature, config.ProfileGuided)
			if err == nil || !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("guided must still refuse %s naming %q, got %v", tc.name, tc.wantSub, err)
			}
		})
	}
}

// TestGuidedStillRefusesUnmetDependencies completes bootstrap first (via a done
// workflow state) so the dependency guard is the ONLY thing left to refuse.
func TestGuidedStillRefusesUnmetDependencies(t *testing.T) {
	coldStart(t, bootstrapPhase(roadmap.Feature{Name: "setup"}),
		roadmap.Phase{Name: "Phase 1", Features: []roadmap.Feature{{Name: "a"}, {Name: "b", DependsOn: []string{"a"}}}})
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	workflow.Save(&workflow.Workflow{Feature: "setup", CurrentStep: "done",
		Steps: map[string]workflow.StepState{}}) //nolint:errcheck
	_, err := workflowOrderForFeature("b", config.ProfileGuided)
	if err == nil || !strings.Contains(err.Error(), "unmet dependencies: a") {
		t.Fatalf("guided must still refuse an unmet dependency naming it, got %v", err)
	}
}

// TestGuidedStillRequiresRoadmapJSON: the one artifact guided never waives.
func TestGuidedStillRequiresRoadmapJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	_, err := workflowOrderForFeature("anything", config.ProfileGuided)
	if err == nil || !strings.Contains(err.Error(), "roadmap.json") {
		t.Fatalf("guided must still refuse without roadmap.json, got %v", err)
	}
}

// TestGuidedDraftRefusalStillSaysDraft: the message the acceptance suite asserts.
func TestGuidedDraftRefusalStillSaysDraft(t *testing.T) {
	coldStart(t, bootstrapPhase(roadmap.Feature{Name: "setup", Draft: true}))
	_, _, err := resolveArchetypeOrder("setup", "", config.ProfileGuided)
	if err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("a draft must still be refused with a message saying \"draft\", got %v", err)
	}
}
