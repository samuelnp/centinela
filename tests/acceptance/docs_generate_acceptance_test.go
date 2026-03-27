package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/generate-html-project-docs.feature
func TestDocsGenerateIncludesMermaidAndValidationWiring(t *testing.T) {
	renderPath := filepath.Join("..", "..", "internal", "docgen", "render.go")
	renderData, err := os.ReadFile(renderPath)
	if err != nil {
		t.Fatalf("read render file: %v", err)
	}
	if !strings.Contains(string(renderData), "Mermaid: Feature Dependencies") {
		t.Fatal("expected mermaid section label")
	}
	validatePath := filepath.Join("..", "..", "cmd", "centinela", "docs_validate.go")
	validateData, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("read docs validate file: %v", err)
	}
	if !strings.Contains(string(validateData), "Documentation inputs are valid") {
		t.Fatal("expected docs validate success output")
	}
}
