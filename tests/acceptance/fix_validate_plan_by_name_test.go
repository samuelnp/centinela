package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: Plan file exists with correct name but minimal content → passes
func TestValidatePlan_CorrectNameMinimalContent_Passes(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.MkdirAll(filepath.Join(dir, "docs/plans"), 0755) //nolint:errcheck
	os.MkdirAll(filepath.Join(dir, "specs"), 0755)      //nolint:errcheck
	os.WriteFile("docs/plans/project-bootstrap.md", []byte("# Plan"), 0644) //nolint:errcheck
	os.WriteFile("specs/a.feature", []byte("Feature: a"), 0644)             //nolint:errcheck

	cfg := &config.Config{}
	if err := workflow.ValidateArtifacts("project-bootstrap", "plan", cfg); err != nil {
		t.Errorf("expected pass, got: %v", err)
	}
}

// Scenario: Plan file is missing entirely → fails with clear error
func TestValidatePlan_MissingFile_FailsWithMessage(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.MkdirAll(filepath.Join(dir, "specs"), 0755)              //nolint:errcheck
	os.WriteFile("specs/a.feature", []byte("Feature: a"), 0644) //nolint:errcheck

	cfg := &config.Config{}
	err := workflow.ValidateArtifacts("missing-feature", "plan", cfg)
	if err == nil {
		t.Error("expected error for missing plan file")
	}
}
