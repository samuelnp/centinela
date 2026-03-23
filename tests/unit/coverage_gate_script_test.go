package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repoRoot() string {
	return filepath.Join("..", "..")
}

func TestCoverageScript_PassesAtOrAboveThreshold(t *testing.T) {
	cmd := exec.Command("sh", "-c", "./scripts/check-coverage.sh")
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), "COVERAGE_VALUE=95.0", "MIN_COVERAGE=95.0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("expected pass, err=%v out=%s", err, out)
	}
}

func TestCoverageScript_UsesDefaultThreshold(t *testing.T) {
	cmd := exec.Command("sh", "-c", "./scripts/check-coverage.sh")
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), "COVERAGE_VALUE=95.0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("expected pass at default threshold, err=%v out=%s", err, out)
	}
}
