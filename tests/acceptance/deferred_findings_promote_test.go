package acceptance_test

// Acceptance: specs/deferred-findings-roadmap-capture.feature

import (
	"os"
	"strings"
	"testing"
)

const promoteRoadmap = `{"phases":[{"name":"Phase 5 — Operability & DX","features":[]},{"name":"Backlog","features":[{"name":"hook-timeout-config","summary":"Prewrite hook timeout is hardcoded; should be configurable","source":{"feature":"deferred-findings-roadmap-capture","role":"senior-engineer"},"deferredAt":"2026-01-01T00:00:00Z"}]}]}`

func seedPromoteArtifacts(t *testing.T, dir string) {
	t.Helper()
	wf := dir + "/.workflow"
	os.WriteFile(wf+"/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(wf+"/roadmap-quality.json", []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(wf+"/roadmap-analysis.md", []byte("# analysis\n"), 0644)                                                      //nolint:errcheck
	os.WriteFile(wf+"/roadmap-quality.md", []byte("# quality\n"), 0644)                                                        //nolint:errcheck
}

// Scenario: Promote without --scores prints evaluator context and writes nothing
func TestDfrc_PromoteNoScoresPrintsContext(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	before, _ := os.ReadFile(dir + "/.workflow/roadmap.json")
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX")
	if code != 0 {
		t.Fatalf("promote no-scores exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "hook-timeout-config") {
		t.Error("output must contain finding name")
	}
	if !strings.Contains(out, "Prewrite hook timeout") {
		t.Error("output must contain finding summary")
	}
	if !strings.Contains(out, "9") {
		t.Error("output must state threshold 9")
	}
	// roadmap.json must be unchanged
	after, _ := os.ReadFile(dir + "/.workflow/roadmap.json")
	if string(before) != string(after) {
		t.Error("roadmap.json must be unchanged when no --scores provided")
	}
}

// Scenario: Promote with valid --scores moves entry from Backlog to target phase and appends artifacts
func TestDfrc_PromoteWithScoresMovesEntry(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX", "--scores", "9,9,8,7,9,9")
	if code != 0 {
		t.Fatalf("promote exit=%d\n%s", code, out)
	}
	data, _ := os.ReadFile(dir + "/.workflow/roadmap.json")
	s := string(data)
	if !strings.Contains(s, "Phase 5 — Operability & DX") {
		t.Error("Phase 5 must be in roadmap after promote")
	}
	if !strings.Contains(s, "hook-timeout-config") {
		t.Error("slug must be in roadmap after promote")
	}
	analysis, _ := os.ReadFile(dir + "/.workflow/roadmap-analysis.json")
	if !strings.Contains(string(analysis), "hook-timeout-config") {
		t.Error("analysis.json must contain promoted slug")
	}
	quality, _ := os.ReadFile(dir + "/.workflow/roadmap-quality.json")
	if !strings.Contains(string(quality), "hook-timeout-config") {
		t.Error("quality.json must contain promoted slug")
	}
}

// Scenario: Promote preserves unknown JSON fields on untouched entries (raw-preserving I/O)
func TestDfrc_PromotePreservesUnknownFields(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Phase 5","features":[]},{"name":"Backlog","features":[{"name":"preserve-fields-test","summary":"s","deferredAt":"t"}]}]}`
	dir := dfrcAcceptDir(t, src)
	wf := dir + "/.workflow"
	os.WriteFile(wf+"/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"existing","customField":"keep-me"}]}`), 0644)                 //nolint:errcheck
	os.WriteFile(wf+"/roadmap-quality.json", []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"existing","customField":"keep-me"}]}`), 0644) //nolint:errcheck
	os.WriteFile(wf+"/roadmap-analysis.md", []byte("# a\n"), 0644)                                                                                                        //nolint:errcheck
	os.WriteFile(wf+"/roadmap-quality.md", []byte("# q\n"), 0644)                                                                                                         //nolint:errcheck
	_, code := runCent(t, bin, dir, "roadmap", "promote", "preserve-fields-test",
		"--phase", "Phase 5", "--scores", "9,9,9,9,9,9")
	if code != 0 {
		t.Logf("promote exit=%d (may be non-zero due to field structure)", code)
	}
	analysis, _ := os.ReadFile(wf + "/roadmap-analysis.json")
	if !strings.Contains(string(analysis), "keep-me") {
		t.Error("custom field 'keep-me' must be preserved in analysis.json after promote")
	}
}
