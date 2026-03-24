package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/edge-case-subagent-tests-phase.feature
func TestEdgeCaseSubagentPrompt_DocIncludesRequiredSections(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "architecture", "edge-case-tester-prompt.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing edge-case prompt doc: %v", err)
	}
	s := string(data)
	checks := []string{"Risk Matrix", "Missing or Weak Scenarios", "Proposed/Added Tests", "Residual Risks", ".workflow/<feature-name>-edge-cases.md"}
	for _, c := range checks {
		if !strings.Contains(s, c) {
			t.Fatalf("prompt doc missing %q", c)
		}
	}
}
