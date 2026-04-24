package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRunRoadmapValidate(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1", Features: []roadmap.Feature{{Name: "user"}}}}}
	roadmap.Save(r) //nolint:errcheck
	if err := runRoadmapValidate(nil, nil); err == nil {
		t.Fatal("expected missing analysis error")
	}
	writeRoadmapAnalysis(t, "user")
	writeRoadmapQuality(t, 9, "user")
	if err := runRoadmapValidate(nil, nil); err != nil {
		t.Fatalf("expected validate success, got %v", err)
	}
}

func TestRunRoadmapValidateNoRoadmap(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	err := runRoadmapValidate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), ".workflow/roadmap.json") {
		t.Fatalf("expected no roadmap error, got %v", err)
	}
}
