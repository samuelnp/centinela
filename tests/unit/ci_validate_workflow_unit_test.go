package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIWorkflowFileExists(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "validate.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected workflow file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "name: Validate") {
		t.Fatal("workflow should declare Validate name")
	}
}
