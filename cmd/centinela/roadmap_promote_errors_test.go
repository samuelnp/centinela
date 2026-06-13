package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestReportPromoteResult_NoRoadmap errors when roadmap.json missing.
func TestReportPromoteResult_NoRoadmap(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)           //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	// No roadmap.json — Load() will fail
	err := reportPromoteResult("some-slug")
	if err == nil {
		t.Fatal("expected error when roadmap.json missing")
	}
	if !strings.Contains(err.Error(), "roadmap") {
		t.Errorf("error must mention roadmap: %v", err)
	}
}

// TestReportPromoteResult_ValidateFails errors when analysis missing.
func TestReportPromoteResult_ValidateFails(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)           //nolint:errcheck
	os.Chdir(d)                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	// Roadmap with the slug in a phase, but no analysis coverage
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 5", Features: []roadmap.Feature{{Name: "some-slug"}}},
	}}
	roadmap.Save(r) //nolint:errcheck
	// No analysis files -> ValidateAnalysis will fail
	err := reportPromoteResult("some-slug")
	if err == nil {
		t.Fatal("expected validate error when analysis missing")
	}
	if !strings.Contains(err.Error(), "validate") {
		t.Errorf("error must mention validate: %v", err)
	}
}

// TestPromoteScored_ParseError errors on bad scores.
func TestPromoteScored_ParseError(t *testing.T) {
	setupPromoteCmd(t)
	promotePhase = "Phase 5"
	promoteScores = "not,valid,scores,at,all,x"
	if err := promoteScored("my-finding"); err == nil {
		t.Fatal("expected error for invalid scores")
	}
}
