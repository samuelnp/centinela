package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/automate-semver-release.feature
func TestVersionBumpWorkflowCommitsAndTagsReleaseVersion(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "version-bump.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)
	checks := []string{
		"chore(release): bump version to $VER [skip ci]",
		"git tag \"v$VER\"",
		"git push origin HEAD:main --tags",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}
