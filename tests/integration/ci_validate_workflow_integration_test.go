package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIWorkflowRunsTestsAndCentinelaValidate(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "validate.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)
	checks := []string{
		"on:",
		"pull_request:",
		"push:",
		"go test ./...",
		"go run ./cmd/centinela validate",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}
