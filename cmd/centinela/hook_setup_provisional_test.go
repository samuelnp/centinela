package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// seededStrictTree builds the exact tree the cross-surface hole needed: a
// STRICT project whose grading artifacts exist only because a guided
// `roadmap promote` seeded them, so they carry "provisional": true.
func seededStrictTree(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	os.WriteFile("ROADMAP.md", []byte("x"), 0644) //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json",
		[]byte(`{"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}]}`), 0644) //nolint:errcheck
	os.MkdirAll("docs/architecture", 0755)                                              //nolint:errcheck
	os.WriteFile("docs/architecture/production-readiness-prompt.md", []byte("x"), 0644) //nolint:errcheck
	if err := roadmap.SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	os.WriteFile("centinela.toml",
		[]byte("[workflow]\nenforcement_profile = \"strict\"\n"), 0644) //nolint:errcheck
}

// TestStrictSetupCascadeRefusesSeededArtifacts: the rungs tested file EXISTENCE
// only, so seeded artifacts satisfied `hook setup` on the very tree where
// `start` refuses with the provisional message. Both surfaces ask the same
// question now.
func TestStrictSetupCascadeRefusesSeededArtifacts(t *testing.T) {
	seededStrictTree(t)
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "roadmap analysis required") {
			t.Fatalf("strict setup must still demand a real senior-PM pass: %q", out)
		}
		if strings.Contains(out, "roadmap checkpoint") {
			t.Fatalf("seeded artifacts must not carry the cascade to the checkpoint: %q", out)
		}
	})
	// The start guard on the identical tree must agree.
	if _, err := workflowOrderForFeature("setup", "strict"); err == nil ||
		!strings.Contains(err.Error(), "provisional") {
		t.Fatalf("start must refuse the same tree as provisional, got %v", err)
	}
}

// TestStrictSetupCascadeAcceptsEvaluatedArtifacts is the ✅ counterweight:
// clearing the mark — what a real evaluator pass does — must carry the cascade
// through, or the rung would be refusing everything rather than refusing seeds.
func TestStrictSetupCascadeAcceptsEvaluatedArtifacts(t *testing.T) {
	seededStrictTree(t)
	for _, p := range []string{roadmap.RoadmapAnalysisFile, roadmap.RoadmapQualityFile} {
		body, _ := os.ReadFile(p)
		cleaned := strings.Replace(string(body), "  \"provisional\": true,\n", "", 1)
		os.WriteFile(p, []byte(cleaned), 0644) //nolint:errcheck
	}
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if strings.Contains(out, "roadmap analysis required") {
			t.Fatalf("an evaluated artifact must satisfy the rung: %q", out)
		}
	})
}

// TestGuidedSetupCascadeStillAdvises: under guided the rungs were never
// blocking, and a seeded artifact must not turn the advisory into a halt.
func TestGuidedSetupCascadeStillAdvises(t *testing.T) {
	seededStrictTree(t)
	os.Remove("centinela.toml") //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if strings.Contains(out, "CENTINELA DIRECTIVE: roadmap analysis required") {
			t.Fatalf("guided must advise, not halt: %q", out)
		}
		if !strings.Contains(out, "roadmap checkpoint") {
			t.Fatalf("guided must still reach the checkpoint: %q", out)
		}
	})
}
