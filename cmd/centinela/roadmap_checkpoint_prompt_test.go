package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// chdirIntoTemp moves into a fresh temp dir and restores cwd on cleanup.
func chdirIntoTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	return d
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// roadmapJSON with a single Phase 0 bootstrap phase containing the named features.
func roadmapJSON(features ...string) string {
	var b strings.Builder
	b.WriteString(`{"phases":[{"name":"Phase 0: Bootstrap","features":[`)
	for i, f := range features {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"name":"` + f + `"}`)
	}
	b.WriteString(`]}]}`)
	return b.String()
}

// layRoadmapArtifacts writes the full set of roadmap-defining artifacts plus
// the production-readiness prompt, so runHookSetup reaches the checkpoint branch.
func layRoadmapArtifacts(t *testing.T, roadmapBody string) {
	t.Helper()
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, "ROADMAP.md", "x")
	writeFile(t, ".workflow/roadmap.json", roadmapBody)
	writeFile(t, ".workflow/roadmap-analysis.md", "x")
	writeFile(t, ".workflow/roadmap-analysis.json", "{}")
	writeFile(t, ".workflow/roadmap-quality.md", "x")
	writeFile(t, ".workflow/roadmap-quality.json", "{}")
	writeFile(t, "docs/architecture/production-readiness-prompt.md", "x")
}

func runSetup(t *testing.T) string {
	t.Helper()
	var out string
	withStdin(t, "{}", func() {
		out = captureStdout(t, func() {
			if err := runHookSetup(nil, nil); err != nil {
				t.Fatalf("runHookSetup returned error: %v", err)
			}
		})
	})
	return out
}

const ckptDirective = "CENTINELA DIRECTIVE: roadmap checkpoint"

func assertContains(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, out)
	}
}

func assertNotContains(t *testing.T, out, notWant string) {
	t.Helper()
	if strings.Contains(out, notWant) {
		t.Fatalf("expected output NOT to contain %q, got:\n%s", notWant, out)
	}
}

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

// Regression / anti-spam: after `centinela roadmap iterate` writes a fresh
// marker, a second runHookSetup with unchanged disk stays SILENT.
//
// NOTE on the mtime-granularity trade-off: WriteMarker stores "at" at RFC3339
// SECOND precision, while filesystem mtimes carry sub-second precision. If the
// marker is written in the SAME wall-clock second as the artifacts, a sub-second
// artifact mtime is After() the second-truncated "at" and the prompt re-fires.
// In the real world the user reviews the roadmap and runs `iterate` seconds or
// minutes later, so this is satisfied. To make the assertion deterministic we
// backdate every artifact mtime to a whole second strictly before the marker is
// written — exercising the real runRoadmapIterate -> WriteMarker path.
func TestCheckpoint_AntiSpam_IterateThenSilent(t *testing.T) {
	chdirIntoTemp(t)
	layRoadmapArtifacts(t, roadmapJSON("phase-0-feature-a"))

	// First run emits.
	out1 := runSetup(t)
	assertContains(t, out1, ckptDirective)

	// Backdate artifacts to a whole second well before "now" so the marker
	// written by iterate is unambiguously >= the latest artifact mtime.
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
