package main

import (
	"os"
	"strings"
	"testing"
)

// TestCov2RunInitSurfacesWorktreeSyncError drives runInit's syncWorktreeIgnores
// error branch: scaffolding succeeds, then a tsconfig.json present as a
// directory makes the ignore sync fail.
func TestCov2RunInitSurfacesWorktreeSyncError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("tsconfig.json", 0o755); err != nil {
		t.Fatal(err)
	}
	prev := agentFlag
	agentFlag = "claude"
	t.Cleanup(func() { agentFlag = prev })
	if err := runInit(nil, nil); err == nil || !strings.Contains(err.Error(), "worktree ignore sync failed") {
		t.Fatalf("expected a worktree sync error, got %v", err)
	}
}

// TestCov2RunStartSaveError drives the workflow.Save error branch: .workflow
// exists as a read-only directory so the state file write is denied.
func TestCov2RunStartSaveError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory enforcement does not apply to root")
	}
	d := t.TempDir()
	t.Chdir(d)
	if err := os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(".workflow", 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(".workflow", 0o755) })
	if err := runStart(nil, []string{"okfeat"}); err == nil || !strings.Contains(err.Error(), "cannot save") {
		t.Fatalf("expected a workflow save error, got %v", err)
	}
}
