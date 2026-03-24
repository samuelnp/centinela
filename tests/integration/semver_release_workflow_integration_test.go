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
		"github.actor != 'github-actions[bot]'",
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
	releaseChecks := []string{"tags:", "- \"v*\"", "GOOS", "GOARCH", "SHA256SUMS", "softprops/action-gh-release@v2"}
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
