package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/improve-docs-llm-hybrid-ui.feature
func TestHybridPromptAndPolishedSectionsExist(t *testing.T) {
	promptPath := filepath.Join("..", "..", "docs", "architecture", "documentation-generator-prompt.md")
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read prompt file: %v", err)
	}
	prompt := string(promptData)
	if !strings.Contains(prompt, "LLM") || !strings.Contains(prompt, "fallback") {
		t.Fatal("prompt should describe llm-first with fallback")
	}
	sectionsPath := filepath.Join("..", "..", "internal", "docgen", "render_sections.go")
	sectionsData, err := os.ReadFile(sectionsPath)
	if err != nil {
		t.Fatalf("read sections file: %v", err)
	}
	sections := string(sectionsData)
	for _, want := range []string{"Feature Topology", "id=\"feature-graphs\""} {
		if !strings.Contains(sections, want) {
			t.Fatalf("missing %q", want)
		}
	}
	helpersPath := filepath.Join("..", "..", "internal", "docgen", "render_helpers.go")
	helpersData, err := os.ReadFile(helpersPath)
	if err != nil {
		t.Fatalf("read helpers file: %v", err)
	}
	helpers := string(helpersData)
	for _, want := range []string{"Documentation Examples", "id=\"examples\""} {
		if !strings.Contains(helpers, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
