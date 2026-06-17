package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: Multiple problems in one run are all reported in a single pass
func TestDoctorMultipleProblemsSinglePass(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".claude/settings.json", "{}") // hooks ERROR
	seedRoadmap(t, dir, "Phase 1: Core")
	writeFile(t, dir, "ROADMAP.md", "drifted\n")                 // roadmap ERROR
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}") // evidence ERROR
	out, code := runDoctor(t, dir)
	if code != 1 {
		t.Fatalf("multiple errors must exit 1, got %d\n%s", code, out)
	}
	for _, name := range []string{"✗ hooks", "✗ roadmap", "✗ evidence"} {
		if !strings.Contains(out, name) {
			t.Fatalf("missing finding %q in single-pass output:\n%s", name, out)
		}
	}
	if !strings.Contains(out, "3 error") {
		t.Fatalf("summary must reflect all three errors:\n%s", out)
	}
}

// Scenario: --fix with multiple fixable problems repairs all of them in one invocation
func TestDoctorFixMultipleInOnePass(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".claude/settings.json", "{}")
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}")
	out, code := runDoctor(t, dir, "--fix")
	if code != 0 {
		t.Fatalf("all-fixable --fix must exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "✓ hooks") || !strings.Contains(out, "✓ evidence") {
		t.Fatalf("both must be repaired in one pass:\n%s", out)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".workflow", "*.json.tmp")); len(left) != 0 {
		t.Fatalf("tmp files must be swept, left %v", left)
	}
}
