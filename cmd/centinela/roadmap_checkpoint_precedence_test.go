package main

import (
	"os"
	"testing"
	"time"
)

// Scenario 8: Precedence — missing ROADMAP.md emits roadmap-required, not checkpoint.
func TestCheckpoint_Precedence_MissingRoadmap(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	out := runSetup(t)
	assertContains(t, out, "CENTINELA DIRECTIVE: roadmap required")
	assertNotContains(t, out, ckptDirective)
}

// Scenario 9: Precedence — invalid roadmap.json yields roadmap json directive.
func TestCheckpoint_Precedence_InvalidRoadmapJSON(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, "ROADMAP.md", "x")
	writeFile(t, ".workflow/roadmap.json", "{not valid json")
	out := runSetup(t)
	assertContains(t, out, "CENTINELA DIRECTIVE: roadmap json")
	assertNotContains(t, out, ckptDirective)
}

// Scenario 10: Multiple Phase 0 features, only the first is done -> picks second.
func TestCheckpoint_MultiFeature_PicksSecond(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-alpha", "phase-0-beta"))
	writeFile(t, ".workflow/phase-0-alpha.json",
		`{"feature":"phase-0-alpha","currentStep":"done","steps":{}}`)
	out := runSetup(t)
	assertContains(t, out, ckptDirective)
	assertContains(t, out, "phase-0-beta")
	assertNotContains(t, out, "centinela start phase-0-alpha")
}

// Scenario 11: Malformed marker JSON is treated as missing and re-emits without crashing.
func TestCheckpoint_MalformedMarker_ReEmits(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	writeFile(t, ".workflow/roadmap-checkpoint.json", "{not valid json")
	out := runSetup(t) // runSetup fails the test if runHookSetup errors/panics.
	assertContains(t, out, ckptDirective)
}

// Scenario 12: Marker "at" unparseable as RFC3339 is treated as stale and re-emits.
func TestCheckpoint_UnparseableAt_ReEmits(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))
	writeFile(t, ".workflow/roadmap-checkpoint.json", `{"choice":"iterate","at":"yesterday"}`)
	out := runSetup(t)
	assertContains(t, out, ckptDirective)
}

// Regression / anti-spam: after `centinela roadmap iterate` writes a fresh marker,
// a second runHookSetup with unchanged disk stays SILENT. Artifacts are backdated
// to a whole second before the marker is written to sidestep the documented
// mtime-granularity trade-off (second-precision "at" vs sub-second mtimes).
func TestCheckpoint_AntiSpam_IterateThenSilent(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))

	// First run emits.
	out1 := runSetup(t)
	assertContains(t, out1, ckptDirective)

	older := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	for _, p := range []string{
		"ROADMAP.md", ".workflow/roadmap.json",
		".workflow/roadmap-analysis.md", ".workflow/roadmap-analysis.json",
		".workflow/roadmap-quality.md", ".workflow/roadmap-quality.json",
	} {
		if err := os.Chtimes(p, older, older); err != nil {
			t.Fatal(err)
		}
	}

	// User chooses "iterate": the real subcommand writes the marker (at = now).
	if err := runRoadmapIterate(nil, nil); err != nil {
		t.Fatalf("runRoadmapIterate: %v", err)
	}
	out2 := runSetup(t)
	assertNotContains(t, out2, ckptDirective)
	assertNotContains(t, out2, "CHECKPOINT")
}
