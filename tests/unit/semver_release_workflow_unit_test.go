package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionBumpWorkflowExists(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "version-bump.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected version bump workflow: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "name: Version Bump") {
		t.Fatal("workflow should declare Version Bump name")
	}
}
