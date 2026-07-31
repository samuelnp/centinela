package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docs-latest-features-getting-started.feature
// Scenario: Getting started teaches the enforced workflow
// The generated-docs story (internal/docgen) was deleted by
// docs-step-markdown-first; the getting-started guide is now the single
// source for onboarding prose and must teach the current markdown-first
// docs-step contract, not the deleted generator commands.
func TestDocsLatestFeaturesAndGettingStartedStayInSync(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "guides", "getting-started.md"))
	if err != nil {
		t.Fatalf("read getting-started guide: %v", err)
	}
	guide := string(data)
	for _, want := range []string{
		"Getting Started",
		"centinela roadmap validate",
		"centinela docs context",
		"changelog",
	} {
		if !strings.Contains(guide, want) {
			t.Fatalf("getting-started guide missing %q", want)
		}
	}
	for _, banned := range []string{"centinela docs generate", "centinela docs validate", "docs/project-docs"} {
		if strings.Contains(guide, banned) {
			t.Fatalf("getting-started guide still teaches deleted %q", banned)
		}
	}
}
