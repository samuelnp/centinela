package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission-denied paths are unreachable as root")
	}
}

// appendIgnoreLine must surface ensureFile's error when a parent path
// component is a regular file (Stat returns a non-NotExist error).
func TestAppendIgnoreLine_ParentIsFile_EnsureErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "blocker")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := appendIgnoreLine(filepath.Join(file, "child"), ignoreEntry); err == nil {
		t.Fatal("appendIgnoreLine must error when a parent component is a file")
	}
}

// appendIgnoreLine must surface a ReadFile error when the target path is a
// directory (ensureFile sees it exists, then ReadFile fails).
func TestAppendIgnoreLine_PathIsDir_ReadErrors(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := appendIgnoreLine(sub, ignoreEntry); err == nil {
		t.Fatal("appendIgnoreLine must error when reading a directory")
	}
}

// appendIgnoreLine must surface the WriteFile error when the existing file is
// read-only and the line is absent.
func TestAppendIgnoreLine_ReadOnlyFile_WriteErrors(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("foo"), 0o444); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := appendIgnoreLine(path, ignoreEntry); err == nil {
		t.Fatal("appendIgnoreLine must error writing a read-only file")
	}
}

// patchTsconfigExclude must surface the WriteFile error when the tsconfig is
// valid, needs patching, but is read-only.
func TestPatchTsconfigExclude_ReadOnly_WriteErrors(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "tsconfig.json")
	if err := os.WriteFile(path, []byte(`{"exclude":["a"]}`), 0o444); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := patchTsconfigExclude(path, ".worktrees"); err == nil {
		t.Fatal("patchTsconfigExclude must error writing a read-only file")
	}
}
