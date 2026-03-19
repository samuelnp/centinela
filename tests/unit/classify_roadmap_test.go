package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestClassifyFile_RoadmapArtifacts(t *testing.T) {
	cfg := &config.Config{}
	cases := []string{
		"/project/docs/features/caesar-cipher.md",
		"/project/ROADMAP.md",
		"/project/.workflow/roadmap.json",
	}
	for _, path := range cases {
		got := workflow.ClassifyFile(path, cfg)
		if got != workflow.TypeRoadmap {
			t.Errorf("ClassifyFile(%q) = %q, want %q", path, got, workflow.TypeRoadmap)
		}
	}
}

func TestIsAllowedInStep_RoadmapAlwaysAllowed(t *testing.T) {
	for _, step := range []string{"plan", "code", "validate"} {
		if !workflow.IsAllowedInStep(workflow.TypeRoadmap, step) {
			t.Errorf("TypeRoadmap should be allowed in %q step", step)
		}
	}
}
