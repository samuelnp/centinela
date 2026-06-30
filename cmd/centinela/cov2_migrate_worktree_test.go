package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCov2MigrateAppliesWorktreeIgnoreSync drives the use_worktrees branch of
// runMigrate: with worktrees enabled and pending setup changes, --apply syncs
// the worktree ignore files after the managed assets land.
func TestCov2MigrateAppliesWorktreeIgnoreSync(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[workflow]\nuse_worktrees = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)

	prevApply, prevAgent := applyFullMigration, fullAgent
	applyFullMigration, fullAgent = true, "both"
	t.Cleanup(func() { applyFullMigration, fullAgent = prevApply, prevAgent })

	if err := runMigrate(nil, nil); err != nil {
		t.Fatalf("runMigrate --apply must succeed, got %v", err)
	}
	// The worktree ignore sync should have created at least .gitignore with the
	// .worktrees/ entry.
	data, err := os.ReadFile(filepath.Join(d, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to be synced: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected a non-empty .gitignore after worktree ignore sync")
	}
}
