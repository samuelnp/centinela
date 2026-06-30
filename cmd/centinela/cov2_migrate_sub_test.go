package main

import (
	"os"
	"testing"
)

// TestCov2MigrateSetupApplyError drives runMigrateSetup's ApplySync error: a
// regular file at .claude blocks the managed settings write during --apply.
func TestCov2MigrateSetupApplyError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".claude", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	prevApply, prevAgent := applySetupMigration, setupAgent
	applySetupMigration, setupAgent = true, "claude"
	t.Cleanup(func() { applySetupMigration, setupAgent = prevApply, prevAgent })
	if err := runMigrateSetup(nil, nil); err == nil {
		t.Fatal("expected a setup ApplySync error")
	}
}

// TestCov2MigrateDocsApplyError drives runMigrateDocs's migration.Apply error: a
// regular file at docs blocks creation of docs/architecture during --apply.
func TestCov2MigrateDocsApplyError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("docs", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := applyDocsMigration
	applyDocsMigration = true
	t.Cleanup(func() { applyDocsMigration = prev })
	if err := runMigrateDocs(nil, nil); err == nil {
		t.Fatal("expected a docs migration Apply error")
	}
}
