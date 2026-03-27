package main

import (
	"os"
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
	if err := runRoadmapValidate(nil, nil); err != nil {
		t.Fatalf("expected validate success, got %v", err)
	}
}
