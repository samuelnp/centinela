// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// gbdColdTree lays the MINIMUM guided greenfield tree: PROJECT.md plus a
// roadmap json with a bootstrap phase. No ROADMAP.md, no roadmap analysis,
// no quality report — the whole point of the slimmed cascade.
func gbdColdTree(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: greenfield\n")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"),
		`{"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}]}`)
}

// Scenario: A guided greenfield project cold-starts from PROJECT.md and roadmap.json alone
func TestGBD_GuidedColdStartFromProjectAndRoadmapAlone(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	out, code := runCent(t, bin, dir, "start", "setup")
	if code != 0 {
		t.Fatalf("guided cold start must succeed on PROJECT.md + roadmap.json alone: %s", out)
	}
	if !strings.Contains(out, "Current step") || !strings.Contains(out, "plan") {
		t.Fatalf("the workflow must start on the plan step: %s", out)
	}
}

// Scenario: A strict greenfield project still requires the full cascade
func TestGBD_StrictGreenfieldStillRequiresFullCascade(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[workflow]\nenforcement_profile = \"strict\"\n")
	out, code := runCent(t, bin, dir, "start", "setup")
	if code == 0 {
		t.Fatalf("strict must still demand the full cascade, got exit 0: %s", out)
	}
	if !strings.Contains(out, "roadmap") || !strings.Contains(out, "analysis") {
		t.Fatalf("the refusal must name the missing roadmap analysis: %s", out)
	}
}

// Scenario: The setup hook advises rather than halts on a guided project
func TestGBD_SetupHookAdvisesOnGuidedProject(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	out, code := runCent(t, bin, dir, "hook", "setup")
	if code != 0 {
		t.Fatalf("hook setup must not fail on a guided project: %s", out)
	}
	if !strings.Contains(out, "Advisory") {
		t.Fatalf("guided must emit an advisory naming the missing artifacts: %s", out)
	}
	if !strings.Contains(out, "roadmap checkpoint") {
		t.Fatalf("guided must still reach the roadmap checkpoint: %s", out)
	}
}

// Scenario: The setup hook still halts on a strict project
func TestGBD_SetupHookHaltsOnStrictProject(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	// ROADMAP.md present so the pre-roadmap rung passes and strict halts
	// specifically on the missing roadmap analysis, matching the spec's
	// "Given a strict project with no roadmap analysis".
	mustWrite(t, filepath.Join(dir, "ROADMAP.md"), "# Roadmap\n")
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[workflow]\nenforcement_profile = \"strict\"\n")
	out, code := runCent(t, bin, dir, "hook", "setup")
	if code != 0 {
		t.Fatalf("hook setup halts via directive, not exit code, got %d: %s", code, out)
	}
	if !strings.Contains(out, "CENTINELA DIRECTIVE") || !strings.Contains(out, "roadmap analysis required") {
		t.Fatalf("strict must emit the roadmap analysis directive: %s", out)
	}
}
