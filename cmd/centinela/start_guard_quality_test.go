package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestWorkflowOrderForFeatureGreenfieldRequiresQualityThreshold(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}}}}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "setup")
	writeRoadmapQuality(t, 8, "setup")
	_, err := workflowOrderForFeature("setup")
	if err == nil || !strings.Contains(err.Error(), "quality evaluation") {
		t.Fatalf("expected roadmap quality error, got %v", err)
	}
}
