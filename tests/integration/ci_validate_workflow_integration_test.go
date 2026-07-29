package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The full suite now runs INSIDE `go run ./cmd/centinela validate`
// (validate.commands executes the single profiled `go test ./...` run), so
// the workflow no longer needs — and must not have — a bare suite step.
func TestCIWorkflowRunsCentinelaValidate(t *testing.T) {
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
		"go run ./cmd/centinela validate",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
	if strings.Contains(content, "name: Run test suite") {
		t.Fatal("workflow must not keep a bare 'Run test suite' step — centinela validate is the single suite run")
	}
}
