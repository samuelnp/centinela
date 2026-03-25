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
		"if: github.actor != 'github-actions[bot]'",
		"chore(release): bump version to $VER",
		"git tag \"v$VER\"",
		"git push origin HEAD:main --tags",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}

func TestTagPushReleasePublishesArtifactsAndChecksums(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "release.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read release workflow file: %v", err)
	}
	content := string(data)
	checks := []string{
		"workflow_run:",
		"workflows: [\"Version Bump\"]",
		"if [ -z \"$TAG\" ]; then",
		"echo \"skip=true\" >> \"$GITHUB_OUTPUT\"",
		"matrix: {goos: [linux, darwin, windows], goarch: [amd64, arm64]}",
		"SHA256SUMS",
		"action-gh-release",
		"dist/*",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("release workflow missing %q", c)
		}
	}
}

func TestInstallerDownloadsMatchingArtifactAndVerifiesChecksum(t *testing.T) {
	path := filepath.Join("..", "..", "scripts", "install.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read installer script: %v", err)
	}
	content := string(data)
	checks := []string{"BIN=\"centinela-${TAG}-${OS}-${ARCH}\"", "BIN_URL=", "EXPECTED=", "ACTUAL=", "install -m 0755"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("installer script missing %q", c)
		}
	}
}
