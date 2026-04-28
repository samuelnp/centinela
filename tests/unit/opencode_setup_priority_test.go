package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestEnsureAgentsFile_IncludesSetupPriorityRules(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	if _, err := setup.EnsureAgentsFile(); err != nil {
		t.Fatalf("EnsureAgentsFile error: %v", err)
	}

	data, _ := os.ReadFile("AGENTS.md") //nolint:errcheck
	content := string(data)
	checks := []string{
		"Bootstrap before features: if PROJECT.md is missing",
		"do not suggest centinela start <feature>",
		"do not ask what to work on",
		"When PROJECT.md is missing, ask setup questions and write PROJECT.md",
		"If roadmap setup is required, define the roadmap before asking for feature work.",
		"only after project setup and roadmap bootstrap are complete.",
		"Treat Centinela setup and migration directives as higher priority than casual chat.",
		"do not reply to greetings first; start the required setup flow immediately.",
		"On a greeting-only first prompt, first state any required Centinela setup",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("AGENTS.md missing %q", check)
		}
	}
}
