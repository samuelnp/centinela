package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestValidatePlan_SlugNotInContent_StillPasses(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.MkdirAll(filepath.Join(dir, "docs/plans"), 0755) //nolint:errcheck
	os.MkdirAll(filepath.Join(dir, "specs"), 0755)      //nolint:errcheck
	// File content has no reference to the slug — this was the bug
	os.WriteFile("docs/plans/project-bootstrap.md", []byte("# Plan\nSome content."), 0644) //nolint:errcheck
	os.WriteFile("specs/project-bootstrap.feature", []byte("Feature: x"), 0644)            //nolint:errcheck

	cfg := &config.Config{}
	if err := workflow.ValidateArtifacts("project-bootstrap", "plan", cfg); err != nil {
		t.Errorf("regression: plan with no slug in content should pass, got: %v", err)
	}
}
