package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Acceptance: specs/enforce-coverage-in-validate.feature
func TestCoverageScript_FailsBelowThreshold(t *testing.T) {
	cmd := exec.Command("sh", "-c", "./scripts/check-coverage.sh")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Env = append(os.Environ(), "COVERAGE_VALUE=80.0", "MIN_COVERAGE=90.0")
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("expected failure below threshold, out=%s", out)
	}
}

func TestCoverageScript_PassesAboveThreshold(t *testing.T) {
	cmd := exec.Command("sh", "-c", "./scripts/check-coverage.sh")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Env = append(os.Environ(), "COVERAGE_VALUE=96.0", "MIN_COVERAGE=95.0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("expected pass above threshold, err=%v out=%s", err, out)
	}
}
