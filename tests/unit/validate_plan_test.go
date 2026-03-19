package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestValidateArtifacts_PlanFoundByFilename(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.MkdirAll(filepath.Join(dir, "docs/plans"), 0755)  //nolint:errcheck
	os.MkdirAll(filepath.Join(dir, "specs"), 0755)       //nolint:errcheck
	os.WriteFile("docs/plans/my-feature.md", []byte("# Plan\nNo slug here."), 0644) //nolint:errcheck
	os.WriteFile("specs/my-feature.feature", []byte("Feature: x"), 0644)            //nolint:errcheck

	cfg := &config.Config{}
	if err := workflow.ValidateArtifacts("my-feature", "plan", cfg); err != nil {
		t.Errorf("expected pass, got: %v", err)
	}
}

func TestValidateArtifacts_PlanMissingFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.MkdirAll(filepath.Join(dir, "specs"), 0755) //nolint:errcheck
	os.WriteFile("specs/x.feature", []byte("Feature: x"), 0644) //nolint:errcheck

	cfg := &config.Config{}
	if err := workflow.ValidateArtifacts("missing-feature", "plan", cfg); err == nil {
		t.Error("expected error for missing plan file")
	}
}
