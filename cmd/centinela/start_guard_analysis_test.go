package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func writeRoadmapAnalysis(t *testing.T, names ...string) {
	t.Helper()
	type item struct {
		Name      string   `json:"name"`
		DependsOn []string `json:"dependsOn"`
	}
	payload := struct {
		Role     string `json:"role"`
		Features []item `json:"features"`
	}{Role: "senior-product-manager"}
	for _, name := range names {
		payload.Features = append(payload.Features, item{Name: name})
	}
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		t.Fatalf("mkdir workflow: %v", err)
	}
	if err := os.WriteFile(roadmap.RoadmapAnalysisMarkdown, []byte("# ok"), 0644); err != nil {
		t.Fatalf("write analysis md: %v", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal analysis: %v", err)
	}
	if err := os.WriteFile(roadmap.RoadmapAnalysisFile, data, 0644); err != nil {
		t.Fatalf("write analysis json: %v", err)
	}
}

func writeRoadmapQuality(t *testing.T, overall int, names ...string) {
	t.Helper()
	type scores struct {
		AcceptanceCriteria int `json:"acceptanceCriteria"`
		UserValue          int `json:"userValue"`
		DefinitionClarity  int `json:"definitionClarity"`
		Dependencies       int `json:"dependencies"`
		EffortEstimation   int `json:"effortEstimation"`
		Overall            int `json:"overall"`
	}
	type item struct {
		Name    string `json:"name"`
		Scores  scores `json:"scores"`
		Summary string `json:"summary"`
	}
	payload := struct {
		Role      string `json:"role"`
		Threshold int    `json:"threshold"`
		Features  []item `json:"features"`
	}{Role: "roadmap-quality-evaluator", Threshold: 9}
	for _, name := range names {
		s := scores{9, 9, 9, 9, 2, overall}
		payload.Features = append(payload.Features, item{Name: name, Scores: s, Summary: "ok"})
	}
	os.WriteFile(".workflow/roadmap-quality.md", []byte("# ok"), 0644) //nolint:errcheck
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal quality: %v", err)
	}
	os.WriteFile(".workflow/roadmap-quality.json", data, 0644) //nolint:errcheck
}

func TestWorkflowOrderForFeatureGreenfieldRequiresAnalysis(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}}}}
	roadmap.Save(r) //nolint:errcheck
	_, err := workflowOrderForFeature("setup")
	if err == nil || !strings.Contains(err.Error(), "senior PM analysis") {
		t.Fatalf("expected roadmap analysis error, got %v", err)
	}
}
