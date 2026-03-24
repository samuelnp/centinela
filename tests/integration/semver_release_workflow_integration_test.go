package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionBumpWorkflowUsesMainAndConventionalCommits(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "version-bump.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)
	checks := []string{
		"branches: [main]",
		"BREAKING CHANGE",
		"^feat(\\([^)]+\\))?:",
		"VERSION :=",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}
