package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// White-box tests for the internal helpers' edge branches.

func TestEnsureFile_AlreadyExists_NotCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a")
	if _, err := ensureFile(path); err != nil {
		t.Fatalf("first ensure: %v", err)
	}
	created, err := ensureFile(path)
	if err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	if created {
		t.Fatal("ensureFile must return false when file already exists")
	}
}

func TestAppendIgnoreLine_NoTrailingNewline_AppendsCleanly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := writeFileNoNewline(path, "existing-entry"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	changed, err := appendIgnoreLine(path, ".worktrees/")
	if err != nil {
		t.Fatalf("appendIgnoreLine: %v", err)
	}
	if !changed {
		t.Fatal("appendIgnoreLine should report changed=true when line was missing")
	}
}

func TestContainsLine_TrimsWhitespace(t *testing.T) {
	data := []byte("  .worktrees/  \n other\n")
	if !containsLine(data, ".worktrees/") {
		t.Fatal("containsLine should ignore leading/trailing whitespace")
	}
	if containsLine(data, ".not-present") {
		t.Fatal("containsLine matched a missing line")
	}
}

func TestHasExclude_BothBranches(t *testing.T) {
	if !hasExclude([]string{"a", ".worktrees", "b"}, ".worktrees") {
		t.Fatal("hasExclude should find existing entry")
	}
	if hasExclude([]string{"a", "b"}, ".worktrees") {
		t.Fatal("hasExclude should not find missing entry")
	}
}

func TestDecodeExcludes_EmptyAndInvalid(t *testing.T) {
	if got := decodeExcludes(nil); got != nil {
		t.Fatalf("decodeExcludes(nil) = %v, want nil", got)
	}
	if got := decodeExcludes([]byte("not-json")); got != nil {
		t.Fatalf("decodeExcludes(invalid) = %v, want nil", got)
	}
}

func writeFileNoNewline(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
