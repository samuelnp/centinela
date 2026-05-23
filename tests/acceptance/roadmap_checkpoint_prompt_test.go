package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Acceptance: specs/roadmap-checkpoint-prompt.feature
//
// These tests build the real centinela binary once and exec `hook setup`
// against temp-dir fixtures, asserting the 12 scenarios end-to-end.

const checkpointDirective = "CENTINELA DIRECTIVE: roadmap checkpoint"

var (
	binOnce sync.Once
	binPath string
	binErr  error
)

// checkpointBin builds the centinela binary once for the package.
func checkpointBin(t *testing.T) string {
	t.Helper()
	binOnce.Do(func() {
		// Keep the binary outside t.TempDir cleanup so subtests can reuse it.
		stable, err := os.MkdirTemp("", "centinela-ckpt-bin")
		if err != nil {
			binErr = err
			return
		}
		binPath = filepath.Join(stable, "centinela-test")
		repo, _ := os.Getwd() // tests/acceptance
		repoRoot := filepath.Clean(filepath.Join(repo, "..", ".."))
		build := exec.Command("go", "build", "-o", binPath, "./cmd/centinela")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			binErr = err
			binPath = string(out)
		}
	})
	if binErr != nil {
		t.Fatalf("build centinela failed: %v\n%s", binErr, binPath)
	}
	return binPath
}

func aWrite(t *testing.T, path, body string) {
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

func aRoadmapJSON(features ...string) string {
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

// fixture lays a project dir with the full roadmap artifact set and returns it.
func fixture(t *testing.T, roadmapBody string) string {
	t.Helper()
	d := t.TempDir()
	at := func(rel, body string) { aWrite(t, filepath.Join(d, rel), body) }
	at("PROJECT.md", "x")
	at("ROADMAP.md", "x")
	at(".workflow/roadmap.json", roadmapBody)
	at(".workflow/roadmap-analysis.md", "x")
	at(".workflow/roadmap-analysis.json", "{}")
	at(".workflow/roadmap-quality.md", "x")
	at(".workflow/roadmap-quality.json", "{}")
	at("docs/architecture/production-readiness-prompt.md", "x")
	return d
}

// runHookSetupIn execs `centinela hook setup` in dir and returns combined output.
func runHookSetupIn(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command(checkpointBin(t), "hook", "setup")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook setup failed in %s: %v\n%s", dir, err, out)
	}
	return string(out)
}

// mustContain / mustNotContain are shared package helpers
// (see diff_aware_gatekeeper_acceptance_test.go).

// Scenario 1: Happy path emits the directive when no marker exists.
func TestAccept_Checkpoint_Emit(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
	mustContain(t, out, "phase-0-feature-a")
}

// Scenario 2: Suppressed when the marker is fresh against all artifacts.
func TestAccept_Checkpoint_SuppressFreshMarker(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	aWrite(t, filepath.Join(d, ".workflow/roadmap-checkpoint.json"),
		`{"choice":"iterate","at":"`+future+`"}`)
	out := runHookSetupIn(t, d)
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 3: Stale marker re-fires when ROADMAP.md is modified after the marker.
func TestAccept_Checkpoint_StaleRoadmap(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	aWrite(t, filepath.Join(d, ".workflow/roadmap-checkpoint.json"),
		`{"choice":"iterate","at":"`+past+`"}`)
	later := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes(filepath.Join(d, "ROADMAP.md"), later, later); err != nil {
		t.Fatal(err)
	}
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
	mustContain(t, out, "phase-0-feature-a")
}

// Scenario 4: Stale marker re-fires when a supporting artifact is modified after.
func TestAccept_Checkpoint_StaleAnalysisArtifact(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	aWrite(t, filepath.Join(d, ".workflow/roadmap-checkpoint.json"),
		`{"choice":"iterate","at":"`+past+`"}`)
	later := time.Now().Add(10 * time.Minute)
	if err := os.Chtimes(filepath.Join(d, ".workflow/roadmap-analysis.json"), later, later); err != nil {
		t.Fatal(err)
	}
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
}

// Scenario 5: Suppressed when bootstrap is already complete.
func TestAccept_Checkpoint_SuppressBootstrapComplete(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	aWrite(t, filepath.Join(d, ".workflow/phase-0-feature-a.json"),
		`{"feature":"phase-0-feature-a","currentStep":"done","steps":{}}`)
	out := runHookSetupIn(t, d)
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 6: Suppressed when no Phase 0 bootstrap features exist.
func TestAccept_Checkpoint_SuppressNoPhaseZero(t *testing.T) {
	d := fixture(t, `{"phases":[{"name":"Phase 1: Features","features":[{"name":"x"}]}]}`)
	out := runHookSetupIn(t, d)
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 7: Suppressed when the workflow file for the first feature exists.
func TestAccept_Checkpoint_SuppressWorkflowFileExists(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	aWrite(t, filepath.Join(d, ".workflow/phase-0-feature-a.json"),
		`{"feature":"phase-0-feature-a","currentStep":"code","steps":{}}`)
	out := runHookSetupIn(t, d)
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 8: Precedence — missing ROADMAP.md emits roadmap-required.
func TestAccept_Checkpoint_PrecedenceMissingRoadmap(t *testing.T) {
	d := t.TempDir()
	aWrite(t, filepath.Join(d, "PROJECT.md"), "x")
	out := runHookSetupIn(t, d)
	mustContain(t, out, "CENTINELA DIRECTIVE: roadmap required")
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 9: Precedence — invalid roadmap.json yields roadmap json directive.
func TestAccept_Checkpoint_PrecedenceInvalidRoadmapJSON(t *testing.T) {
	d := t.TempDir()
	aWrite(t, filepath.Join(d, "PROJECT.md"), "x")
	aWrite(t, filepath.Join(d, "ROADMAP.md"), "x")
	aWrite(t, filepath.Join(d, ".workflow/roadmap.json"), "{not valid json")
	out := runHookSetupIn(t, d)
	mustContain(t, out, "CENTINELA DIRECTIVE: roadmap json")
	mustNotContain(t, out, checkpointDirective)
}

// Scenario 10: Multiple Phase 0 features, only the first is done -> picks second.
func TestAccept_Checkpoint_MultiFeaturePicksSecond(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-alpha", "phase-0-beta"))
	aWrite(t, filepath.Join(d, ".workflow/phase-0-alpha.json"),
		`{"feature":"phase-0-alpha","currentStep":"done","steps":{}}`)
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
	mustContain(t, out, "phase-0-beta")
}

// Scenario 11: Malformed marker JSON is treated as missing and re-emits without crashing.
func TestAccept_Checkpoint_MalformedMarkerReEmits(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	aWrite(t, filepath.Join(d, ".workflow/roadmap-checkpoint.json"), "{not valid json")
	// runHookSetupIn fails the test on a non-zero exit (crash), satisfying
	// "the command should not crash".
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
}

// Scenario 12: Marker "at" unparseable as RFC3339 is treated as stale and re-emits.
func TestAccept_Checkpoint_UnparseableAtReEmits(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	aWrite(t, filepath.Join(d, ".workflow/roadmap-checkpoint.json"),
		`{"choice":"iterate","at":"yesterday"}`)
	out := runHookSetupIn(t, d)
	mustContain(t, out, checkpointDirective)
}

// Anti-spam regression: after `centinela roadmap iterate` writes a fresh marker,
// a second `hook setup` with unchanged disk stays silent. Exercises the real
// iterate subcommand end-to-end. Artifacts are backdated to a whole second so
// the second-precision marker "at" is unambiguously >= every artifact mtime
// (see the mtime-granularity note in the edge-case report).
func TestAccept_Checkpoint_AntiSpamIterateThenSilent(t *testing.T) {
	d := fixture(t, aRoadmapJSON("phase-0-feature-a"))
	out1 := runHookSetupIn(t, d)
	mustContain(t, out1, checkpointDirective)

	older := time.Now().Add(-time.Hour).Truncate(time.Second)
	for _, p := range []string{
		"ROADMAP.md", ".workflow/roadmap.json",
		".workflow/roadmap-analysis.md", ".workflow/roadmap-analysis.json",
		".workflow/roadmap-quality.md", ".workflow/roadmap-quality.json",
	} {
		if err := os.Chtimes(filepath.Join(d, p), older, older); err != nil {
			t.Fatal(err)
		}
	}

	iterate := exec.Command(checkpointBin(t), "roadmap", "iterate")
	iterate.Dir = d
	if out, err := iterate.CombinedOutput(); err != nil {
		t.Fatalf("roadmap iterate failed: %v\n%s", err, out)
	}
	out2 := runHookSetupIn(t, d)
	mustNotContain(t, out2, checkpointDirective)
}
