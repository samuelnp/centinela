package main

import (
	"os"
	"testing"
	"time"
)

// Scenario 1: Happy path emits the checkpoint directive when no marker exists.
func TestCheckpoint_Emit_NoMarker(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	out := runSetup(t)
	assertContains(t, out, ckptDirective)
	assertContains(t, out, "phase-0-feature-a")
}

// Scenario 2: Suppressed when the marker is fresh against all artifacts.
func TestCheckpoint_Suppressed_FreshMarker(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	// Marker "at" in the future relative to all artifact mtimes.
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	writeFile(t, ".workflow/roadmap-checkpoint.json", `{"choice":"iterate","at":"`+future+`"}`)
	out := runSetup(t)
	assertNotContains(t, out, ckptDirective)
	assertNotContains(t, out, "CHECKPOINT")
}

// Scenario 3: Stale marker re-fires when ROADMAP.md is modified after the marker.
func TestCheckpoint_Stale_RoadmapNewer(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	writeFile(t, ".workflow/roadmap-checkpoint.json", `{"choice":"iterate","at":"`+past+`"}`)
	// ROADMAP.md mtime strictly after the marker "at".
	later := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes("ROADMAP.md", later, later); err != nil {
		t.Fatal(err)
	}
	out := runSetup(t)
	assertContains(t, out, ckptDirective)
	assertContains(t, out, "phase-0-feature-a")
}

// Scenario 4: Stale marker re-fires when a supporting artifact is modified after.
func TestCheckpoint_Stale_AnalysisArtifactNewer(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	writeFile(t, ".workflow/roadmap-checkpoint.json", `{"choice":"iterate","at":"`+past+`"}`)
	later := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes(".workflow/roadmap-analysis.json", later, later); err != nil {
		t.Fatal(err)
	}
	out := runSetup(t)
	assertContains(t, out, ckptDirective)
	assertContains(t, out, "phase-0-feature-a")
}

// Scenario 5: Suppressed when bootstrap is already complete (all features done).
func TestCheckpoint_Suppressed_BootstrapComplete(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	writeFile(t, ".workflow/phase-0-feature-a.json",
		`{"feature":"phase-0-feature-a","currentStep":"done","steps":{}}`)
	out := runSetup(t)
	assertNotContains(t, out, ckptDirective)
}

// Scenario 6: Suppressed when no Phase 0 bootstrap features exist.
func TestCheckpoint_Suppressed_NoPhaseZero(t *testing.T) {
	chdirIntoTemp(t)
	// Roadmap has only a non-bootstrap phase.
	layRoadmapArtifacts(t, `{"phases":[{"name":"Phase 1: Features","features":[{"name":"x"}]}]}`)
	out := runSetup(t)
	assertNotContains(t, out, ckptDirective)
}

// Scenario 7: Suppressed when the workflow file for the first feature exists
// (and is not done) — feature already started.
func TestCheckpoint_Suppressed_WorkflowFileExists(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	// In-progress (not done): FirstIncompleteBootstrap still picks it, but
	// Decide suppresses because the workflow file exists.
	writeFile(t, ".workflow/phase-0-feature-a.json",
		`{"feature":"phase-0-feature-a","currentStep":"code","steps":{}}`)
	out := runSetup(t)
	assertNotContains(t, out, ckptDirective)
	assertNotContains(t, out, "CHECKPOINT")
}
