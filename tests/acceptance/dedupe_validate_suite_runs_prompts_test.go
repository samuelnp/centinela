package acceptance_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/dedupe-validate-suite-runs.feature

func readPromptPair(t *testing.T, name string) (string, string) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "..", "docs", "architecture", name))
	if err != nil {
		t.Fatalf("read source prompt: %v", err)
	}
	mirror, err := os.ReadFile(filepath.Join("..", "..", "internal", "scaffold",
		"assets", "docs", "architecture", name))
	if err != nil {
		t.Fatalf("read scaffold mirror: %v", err)
	}
	if !bytes.Equal(src, mirror) {
		t.Fatalf("%s and its scaffold mirror must be byte-identical", name)
	}
	return string(src), string(mirror)
}

// Scenario: gatekeeper prompt copies stay identical and carry the conditional suite mandate
func TestGatekeeperPromptCopiesCarryConditionalSuiteMandate(t *testing.T) {
	src, _ := readPromptPair(t, "gatekeeper-prompt.md")
	for _, want := range []string{
		"ONLY when",
		"does not already execute it in full",
		"IS the suite",
		"do NOT run the suite a\n   second time",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("gatekeeper prompt missing conditional-mandate wording %q", want)
		}
	}
}

// Scenario: qa-senior prompt copies carry the one-full-run guidance
func TestQASeniorPromptCopiesCarryOneFullRunGuidance(t *testing.T) {
	src, _ := readPromptPair(t, "qa-senior-prompt.md")
	for _, want := range []string{
		"run ONLY the affected packages' tests",
		"exactly ONE full profiled run",
		"go test ./... -coverprofile=coverage.out",
		"COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh",
		"do not repeat it",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("qa-senior prompt missing one-full-run wording %q", want)
		}
	}
}
