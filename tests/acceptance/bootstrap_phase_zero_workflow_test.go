package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/bootstrap-phase-zero-workflow.feature
func TestBootstrapWorkflowUsesDocsStepAndStrictTestArtifacts(t *testing.T) {
	workflowPath := filepath.Join("..", "..", "internal", "workflow", "order.go")
	workflowData, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read order file: %v", err)
	}
	workflowContent := string(workflowData)
	if !strings.Contains(workflowContent, `{"plan", "code", "validate", "docs"}`) {
		t.Fatal("bootstrap step order missing")
	}
	validatePath := filepath.Join("..", "..", "internal", "workflow", "validate_tests.go")
	validateData, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("read validate tests file: %v", err)
	}
	validateContent := string(validateData)
	if !strings.Contains(validateContent, `name != ".gitkeep"`) {
		t.Fatal("placeholder test artifact guard missing")
	}
}
