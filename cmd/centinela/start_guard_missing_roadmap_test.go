package main

import (
	"os"
	"strings"
	"testing"
)

func TestWorkflowOrderForFeatureReportsMissingRoadmapJSON(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("# Roadmap\n"), 0644)                 //nolint:errcheck
	_, err := workflowOrderForFeature("setup")
	if err == nil || !strings.Contains(err.Error(), ".workflow/roadmap.json") || !strings.Contains(err.Error(), "centinela roadmap validate") {
		t.Fatalf("expected explicit roadmap json guidance, got %v", err)
	}
}
