package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestCov2PrecommitUninstallSurfacesReadError drives runPrecommitUninstall's
// error branch: the pre-commit hook present as a directory makes the read fail.
func TestCov2PrecommitUninstallSurfacesReadError(t *testing.T) {
	d := t.TempDir()
	t.Chdir(d)
	if err := os.MkdirAll(filepath.Join(".git", "hooks", "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runPrecommitUninstall(nil, nil); err == nil {
		t.Fatal("expected an uninstall read error")
	}
}

// TestCov2MigrateFullDocsApplyError drives runMigrate's docs migration.Apply
// error during --apply: a regular file at docs blocks docs/architecture writes.
func TestCov2MigrateFullDocsApplyError(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "docs"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	prevApply, prevAgent := applyFullMigration, fullAgent
	applyFullMigration, fullAgent = true, "both"
	t.Cleanup(func() { applyFullMigration, fullAgent = prevApply, prevAgent })
	if err := runMigrate(nil, nil); err == nil {
		t.Fatal("expected a docs migration apply error")
	}
}

// TestCov2RunCompleteSaveError drives complete's saveWorkflow error branch: the
// code step advances without gates, but a read-only .workflow blocks the save.
func TestCov2RunCompleteSaveError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory enforcement does not apply to root")
	}
	d := t.TempDir()
	t.Chdir(d)
	if err := os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Outcome profile keeps the code step free of strict orchestration evidence,
	// so completion reaches the save step rather than a gate failure.
	wf := workflow.NewWithOrder("feat", workflow.DefaultStepOrder, config.ProfileOutcome)
	wf.EnforcementProfile = config.ProfileOutcome
	wf.CurrentStep = "code"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(workflow.FilePath("feat"), 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(workflow.FilePath("feat"), 0o644) })
	err := runComplete(nil, []string{"feat"})
	if err == nil || !strings.Contains(err.Error(), "cannot save") {
		t.Fatalf("expected a save error, got %v", err)
	}
}
