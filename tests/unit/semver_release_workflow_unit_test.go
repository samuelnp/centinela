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
	checks := []string{"name: Version Bump", "branches: [main]", "concurrency:", "group: version-bump-main"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("workflow missing %q", c)
		}
	}
}

func TestReleaseWorkflowAndInstallerExist(t *testing.T) {
	workflowPath := filepath.Join("..", "..", ".github", "workflows", "release.yml")
	workflowData, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("expected release workflow: %v", err)
	}
	if !strings.Contains(string(workflowData), "name: Release") {
		t.Fatal("release workflow should declare Release name")
	}
	installerPath := filepath.Join("..", "..", "scripts", "install.sh")
	installerData, err := os.ReadFile(installerPath)
	if err != nil {
		t.Fatalf("expected installer script: %v", err)
	}
	if !strings.Contains(string(installerData), "#!/usr/bin/env bash") {
		t.Fatal("installer should be a bash script")
	}
}
