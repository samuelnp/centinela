package unit_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestCoverageScript_ReusesExistingProfileWithoutRunningSuite proves the
// reuse branch: with COVERAGE_PROFILE naming a profile recorded by a
// passing run, the script must NOT run the suite — the module now carries
// a failing test, so any internal `go test` re-run would fail under set -e.
func TestCoverageScript_ReusesExistingProfileWithoutRunningSuite(t *testing.T) {
	dir := newTempGoModule(t)
	writeProfile(t, dir)
	addFailingTest(t, dir)

	cmd := exec.Command("sh", coverageScriptAbs(t))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=",
		"COVERAGE_PROFILE=profile.out", "MIN_COVERAGE=0.0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("reuse branch must skip the (now failing) suite, err=%v out=%s", err, out)
	}
	if !strings.Contains(string(out), "coverage gate passed") {
		t.Fatalf("expected pass message, out=%s", out)
	}
}

// TestCoverageScript_MissingProfileFallsBackToRunningSuite proves the
// fail-safe: COVERAGE_PROFILE naming a missing file never skips coverage —
// the script runs the suite itself, and the module's failing test makes
// that run (and so the script) fail. Never fail-open.
func TestCoverageScript_MissingProfileFallsBackToRunningSuite(t *testing.T) {
	dir := newTempGoModule(t)
	addFailingTest(t, dir)

	cmd := exec.Command("sh", coverageScriptAbs(t))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=",
		"COVERAGE_PROFILE=missing.out", "MIN_COVERAGE=0.0")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("missing profile must trigger a real suite run (which fails), out=%s", out)
	}
}
