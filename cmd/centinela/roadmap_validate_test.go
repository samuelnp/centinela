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

// TestRunRoadmapValidateLowScoreIsAdvisory: a feature scored below the deleted
// threshold validates (exit 0) and is named in the advisory summary instead.
func TestRunRoadmapValidateLowScoreIsAdvisory(t *testing.T) {
	t.Chdir(t.TempDir())
	roadmap.Save(&roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1",
		Features: []roadmap.Feature{{Name: "user"}}}}}) //nolint:errcheck
	writeRoadmapAnalysis(t, "user")
	writeRoadmapQuality(t, 3, "user")
	out := captureStdout(t, func() {
		if err := runRoadmapValidate(nil, nil); err != nil {
			t.Fatalf("a low score must not refuse, got %v", err)
		}
	})
	if !strings.Contains(out, "Advisory") || !strings.Contains(out, "user") ||
		!strings.Contains(out, "overall 3") {
		t.Fatalf("expected an advisory naming the feature and its score, got %q", out)
	}
}

// TestRunRoadmapValidateHealthyScoresAreSilent is the ❌ direction: no advisory
// noise on a roadmap that has nothing to advise about.
func TestRunRoadmapValidateHealthyScoresAreSilent(t *testing.T) {
	t.Chdir(t.TempDir())
	roadmap.Save(&roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1",
		Features: []roadmap.Feature{{Name: "user"}}}}}) //nolint:errcheck
	writeRoadmapAnalysis(t, "user")
	writeRoadmapQuality(t, 9, "user")
	out := captureStdout(t, func() { _ = runRoadmapValidate(nil, nil) })
	if strings.Contains(out, "Advisory") {
		t.Fatalf("healthy scores must draw no advisory, got %q", out)
	}
}
