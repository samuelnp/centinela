package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestWorkflowOrderForFeature_BacklogRefused refuses a Backlog feature with
// a promote-first error (regression: start-guard Backlog check).
func TestWorkflowOrderForFeature_BacklogRefused(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                    //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{
		Phases: []roadmap.Phase{
			{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "bootstrap-done"}}},
			{Name: "Backlog", Features: []roadmap.Feature{{Name: "backlog-finding"}}},
		},
	}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "bootstrap-done")
	writeRoadmapQuality(t, 9, "bootstrap-done")

	_, err := workflowOrderForFeature("backlog-finding")
	if err == nil {
		t.Fatal("starting a Backlog feature must be refused")
	}
	if !strings.Contains(err.Error(), "promote") {
		t.Errorf("error must mention promote, got: %v", err)
	}
	if !strings.Contains(err.Error(), "backlog-finding") {
		t.Errorf("error must name the feature, got: %v", err)
	}
}
