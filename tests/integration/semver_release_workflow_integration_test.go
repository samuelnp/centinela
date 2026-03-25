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
		"github.actor != 'github-actions[bot]'",
		"LAST_TAG=$(git tag --list 'v*' --sort=-version:refname | head -n1 || true)",
		"if [ -n \"$LAST_TAG\" ]; then",
		"LOG=$(git log --format='%s%n%b' HEAD || true)",
		"BREAKING CHANGE",
		"^feat(\\([^)]+\\))?:",
		"VERSION :=",
		"git -c user.name=\"github-actions[bot]\" -c user.email=\"41898282+github-actions[bot]@users.noreply.github.com\" commit",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}

func TestReleaseWorkflowAndInstallerContainExpectedFlow(t *testing.T) {
	releasePath := filepath.Join("..", "..", ".github", "workflows", "release.yml")
	releaseData, err := os.ReadFile(releasePath)
	if err != nil {
		t.Fatalf("read release workflow file: %v", err)
	}
	releaseContent := string(releaseData)
	releaseChecks := []string{
		"workflow_run:",
		"workflows: [\"Version Bump\"]",
		"github.event.workflow_run.conclusion == 'success'",
		"SHA=\"${{ github.event.workflow_run.head_sha }}\"",
		"git tag --points-at \"$SHA\"",
		"echo \"skip=true\" >> \"$GITHUB_OUTPUT\"",
		"tag_name: ${{ needs.build-and-release.outputs.tag }}",
		"matrix: {goos: [linux, darwin, windows], goarch: [amd64, arm64]}",
		"SHA256SUMS",
		"softprops/action-gh-release@v2",
	}
	for _, c := range releaseChecks {
		if !strings.Contains(releaseContent, c) {
			t.Fatalf("release workflow missing %q", c)
		}
	}
	installerPath := filepath.Join("..", "..", "scripts", "install.sh")
	installerData, err := os.ReadFile(installerPath)
	if err != nil {
		t.Fatalf("read installer script: %v", err)
	}
	installerContent := string(installerData)
	installerChecks := []string{"uname -s", "uname -m", "releases/latest", "SHA256SUMS", "checksum verification failed"}
	for _, c := range installerChecks {
		if !strings.Contains(installerContent, c) {
			t.Fatalf("installer script missing %q", c)
		}
	}
}
