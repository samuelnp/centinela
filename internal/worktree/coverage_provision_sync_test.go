package worktree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission-denied paths are unreachable as root")
	}
}

// SyncIgnores surfaces appendIgnoreLine's create error when the repo dir is
// read-only (ensureFile cannot write the missing ignore file).
func TestSyncIgnores_ReadOnlyRepoErrors(t *testing.T) {
	skipIfRoot(t)
	repo := t.TempDir()
	if err := os.Chmod(repo, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(repo, 0o755)
	if _, err := worktree.SyncIgnores(repo); err == nil {
		t.Fatal("SyncIgnores must error on a read-only repo")
	}
}

// SyncIgnores surfaces the tsconfig read error when tsconfig.json is a
// directory rather than a file.
func TestSyncIgnores_TsconfigIsDirErrors(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, "tsconfig.json"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := worktree.SyncIgnores(repo); err == nil {
		t.Fatal("SyncIgnores must error when tsconfig.json is a directory")
	}
}

// Create surfaces the MkdirAll error when `.worktrees` is occupied by a file.
func TestCreate_WorktreesParentIsFileErrors(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, ".worktrees"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := worktree.Create(repo, "feat"); err == nil {
		t.Fatal("Create must error when .worktrees is a regular file")
	}
}

// Merge surfaces the isDirty git error when run outside a git repository.
func TestMerge_NonRepoIsDirtyErrors(t *testing.T) {
	repo := t.TempDir()
	run := func(string) (bool, string) { return true, "" }
	if _, err := worktree.Merge(repo, "feat", run); err == nil {
		t.Fatal("Merge must surface the git status error outside a repo")
	}
}
