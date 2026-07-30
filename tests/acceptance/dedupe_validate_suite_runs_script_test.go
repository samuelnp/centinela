package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/dedupe-validate-suite-runs.feature
//
// The reuse and fallback branches are proven BEHAVIORALLY at unit level
// (tests/unit/coverage_reuse_branch_test.go, temp Go module); here the
// script text pins the branch conditions so a rewrite cannot silently
// invert them. No test in this file runs the repo's own suite.

func coverageScriptText(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "scripts", "check-coverage.sh"))
	if err != nil {
		t.Fatalf("read check-coverage.sh: %v", err)
	}
	return string(data)
}

// Scenario: bare check-coverage.sh invocation stays self-contained
func TestBareCoverageInvocationStaysSelfContained(t *testing.T) {
	text := coverageScriptText(t)
	// COVERAGE_PROFILE unset → the -z arm is true, so the script runs the
	// suite itself even when a stale coverage.out sits on disk.
	if !strings.Contains(text, `if [ -z "${COVERAGE_PROFILE:-}" ] || [ ! -f "$PROFILE" ]; then`) {
		t.Fatal("bare invocation must run the suite itself (reuse only when COVERAGE_PROFILE is explicitly set)")
	}
	if !strings.Contains(text, `THRESHOLD="${MIN_COVERAGE:-95.0}"`) {
		t.Fatal("default MIN_COVERAGE floor of 95.0 must be applied")
	}
}

// Scenario: check-coverage.sh reuses an existing profile when COVERAGE_PROFILE is set
func TestReuseBranchGuardedByExplicitProfileAndExistence(t *testing.T) {
	text := coverageScriptText(t)
	if !strings.Contains(text, `PROFILE="${COVERAGE_PROFILE:-coverage.out}"`) {
		t.Fatal("script must resolve the profile from COVERAGE_PROFILE")
	}
	if !strings.Contains(text, `if [ -z "${COVERAGE_PROFILE:-}" ] || [ ! -f "$PROFILE" ]; then`) {
		t.Fatal("reuse requires COVERAGE_PROFILE explicitly set AND the file present")
	}
}

// Scenario: check-coverage.sh falls back to running the suite when the profile is missing
func TestMissingProfileFallsBackToSelfRun(t *testing.T) {
	text := coverageScriptText(t)
	// The [ ! -f "$PROFILE" ] arm makes a missing named profile fail-safe:
	// the suite runs and (re)writes the named profile — never fail-open.
	if !strings.Contains(text, `go test ./... -coverprofile="$PROFILE"`) {
		t.Fatal("fallback must run the suite and write the named profile")
	}
}

// Scenario: COVERAGE_VALUE fast path bypasses the suite unchanged
func TestCoverageValueFastPathBypassesSuite(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "check-coverage.sh"))
	if err != nil {
		t.Fatalf("abs script path: %v", err)
	}
	// Empty temp dir: no go.mod, no profile. Any suite run or profile read
	// would fail here, so exit 0 proves the fast path bypasses both.
	cmd := exec.Command("sh", script)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "COVERAGE_VALUE=99.0", "MIN_COVERAGE=95.0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("COVERAGE_VALUE fast path must pass without suite or profile, err=%v out=%s", err, out)
	}
	if !strings.Contains(string(out), "coverage gate passed") {
		t.Fatalf("expected pass message, out=%s", out)
	}
}
