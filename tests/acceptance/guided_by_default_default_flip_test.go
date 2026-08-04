// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gbdExisting lays PROJECT.md as an "existing" project so the greenfield
// roadmap-grading cascade never engages — these scenarios are about profile
// RESOLUTION, not the cascade (covered in guided_by_default_cascade_test.go).
func gbdExisting(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
}

// gbdStatus runs `centinela status <feature>` and returns its output.
func gbdStatus(t *testing.T, bin, dir, feature string) string {
	t.Helper()
	out, code := runCent(t, bin, dir, "status", feature)
	if code != 0 {
		t.Fatalf("status failed: %s", out)
	}
	return out
}

// Scenario: A new workflow in a zero-config project defaults to guided
func TestGBD_ZeroConfigDefaultsToGuided(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdExisting(t, dir)
	if out, code := runCent(t, bin, dir, "start", "zc-feature"); code != 0 {
		t.Fatalf("start failed: %s", out)
	}
	out := gbdStatus(t, bin, dir, "zc-feature")
	if !strings.Contains(out, "guided") || !strings.Contains(out, "default (guided)") {
		t.Fatalf("zero-config start must resolve to guided as the default: %s", out)
	}
}

// Scenario: A workflow started before the flip keeps strict
func TestGBD_LegacyWorkflowKeepsStrict(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".workflow"), 0755) //nolint:errcheck
	// Hand-written legacy state: no profileContract, no enforcementProfile —
	// exactly the shape of a pre-flip workflow on disk.
	mustWrite(t, filepath.Join(dir, ".workflow", "legacy-wf.json"),
		`{"feature":"legacy-wf","currentStep":"plan","steps":{}}`)
	out := gbdStatus(t, bin, dir, "legacy-wf")
	if !strings.Contains(out, "strict") || !strings.Contains(out, "default (strict, legacy workflow)") {
		t.Fatalf("an unpinned legacy workflow must keep strict, got: %s", out)
	}
}
