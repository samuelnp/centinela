package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/add-ci-validate-workflow.feature
func TestCIWorkflowIncludesCoverageGateViaCentinelaValidate(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "validate.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "go run ./cmd/centinela validate") {
		t.Fatal("workflow must run centinela validate")
	}
}
