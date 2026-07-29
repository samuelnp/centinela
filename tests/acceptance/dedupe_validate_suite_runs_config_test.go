package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// Acceptance: specs/dedupe-validate-suite-runs.feature

// Scenario: validate runs the suite exactly once
func TestValidateCommandsRunSuiteExactlyOnceWithProfileReuse(t *testing.T) {
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                //nolint:errcheck
	os.Chdir(filepath.Join("..", "..")) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load error: %v", err)
	}
	suiteRuns := 0
	var suiteCmd, coverageCmd string
	for _, c := range cfg.Validate.Commands {
		if strings.Contains(c, "go test ./...") {
			suiteRuns++
			suiteCmd = c
		}
		if strings.Contains(c, "scripts/check-coverage.sh") {
			coverageCmd = c
		}
	}
	if suiteRuns != 1 {
		t.Fatalf("exactly one validate command must run the full suite, got %d", suiteRuns)
	}
	if !strings.Contains(suiteCmd, "-coverprofile=coverage.out") {
		t.Fatalf("the single suite run %q must write the coverage profile", suiteCmd)
	}
	if !strings.Contains(coverageCmd, "COVERAGE_PROFILE=coverage.out") {
		t.Fatalf("coverage command %q must reuse the suite run's profile, not re-run it", coverageCmd)
	}
}

// Scenario: CI runs the suite once via centinela validate
func TestCIRunsSuiteOnceViaCentinelaValidate(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "validate.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "name: Run test suite") {
		t.Fatal("validate.yml must not keep a bare 'Run test suite' step")
	}
	if strings.Contains(content, "run: go test ./...") {
		t.Fatal("validate.yml must not run the bare suite outside centinela validate")
	}
	if !strings.Contains(content, "go run ./cmd/centinela validate") {
		t.Fatal("validate.yml must keep centinela validate as the single suite execution")
	}
}
