package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestValidateCommands_CoverageScriptPairsWithProfiledRun asserts the
// single-run design: validate.commands still invokes the coverage script,
// and it PAIRS with the profiled suite run — one command writes
// coverage.out via -coverprofile, and the coverage entry reuses that same
// profile through COVERAGE_PROFILE instead of re-running the suite.
func TestValidateCommands_CoverageScriptPairsWithProfiledRun(t *testing.T) {
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                //nolint:errcheck
	os.Chdir(filepath.Join("..", "..")) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load error: %v", err)
	}
	var coverageCmd string
	profiledRun := false
	for _, c := range cfg.Validate.Commands {
		if strings.Contains(c, "scripts/check-coverage.sh") {
			coverageCmd = c
		}
		if strings.Contains(c, "-coverprofile=coverage.out") {
			profiledRun = true
		}
	}
	if coverageCmd == "" {
		t.Fatal("validate commands should include scripts/check-coverage.sh")
	}
	if !profiledRun {
		t.Fatal("validate commands should include a -coverprofile=coverage.out suite run")
	}
	if !strings.Contains(coverageCmd, "COVERAGE_PROFILE=coverage.out") {
		t.Fatalf("coverage command %q must reuse the profiled run via COVERAGE_PROFILE=coverage.out", coverageCmd)
	}
}
