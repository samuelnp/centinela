package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/generate-html-project-docs.feature
func TestDocsGenerateIncludesFeatureMermaidAndValidationWiring(t *testing.T) {
	renderPath := filepath.Join("..", "..", "internal", "docgen", "render.go")
	renderData, err := os.ReadFile(renderPath)
	if err != nil {
		t.Fatalf("read render file: %v", err)
	}
	if !strings.Contains(string(renderData), "renderFeatureGraphs") {
		t.Fatal("expected feature graph section renderer")
	}
	graphsPath := filepath.Join("..", "..", "internal", "docgen", "render_graphs.go")
	graphsData, err := os.ReadFile(graphsPath)
	if err != nil {
		t.Fatalf("read render graphs file: %v", err)
	}
	if strings.Contains(string(graphsData), "mermaidEvidence") {
		t.Fatal("workflow-specific mermaid graph should not be present")
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
