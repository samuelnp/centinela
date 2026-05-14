package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestPath_JoinsRepoAndFeature(t *testing.T) {
	got := worktree.Path("/repo", "alpha")
	want := filepath.Join("/repo", ".worktrees", "alpha")
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestDetectFeatureFromCwd_Inside(t *testing.T) {
	cwd := filepath.Join(string(filepath.Separator), "repo", ".worktrees", "alpha", "src")
	feat, root := worktree.DetectFeatureFromCwd(cwd)
	if feat != "alpha" {
		t.Fatalf("feature = %q, want %q", feat, "alpha")
	}
	if filepath.ToSlash(root) == "" || filepath.Base(root) != "alpha" {
		t.Fatalf("root = %q, want path ending in alpha", root)
	}
}

func TestDetectFeatureFromCwd_Outside(t *testing.T) {
	cwd := filepath.Join(string(filepath.Separator), "repo", "internal", "x")
	feat, root := worktree.DetectFeatureFromCwd(cwd)
	if feat != "" || root != "" {
		t.Fatalf("expected outside, got feature=%q root=%q", feat, root)
	}
}

func TestDetectFeatureFromCwd_NoFeatureAfterDir(t *testing.T) {
	// Exactly `.worktrees` with no child returns no feature.
	cwd := filepath.Join(string(filepath.Separator), "repo", ".worktrees")
	feat, _ := worktree.DetectFeatureFromCwd(cwd)
	if feat != "" {
		t.Fatalf("expected empty feature at .worktrees only, got %q", feat)
	}
}

func TestDetectFeatureFromCwd_WithSymlinks(t *testing.T) {
	// Mirror the macOS /tmp → /private/tmp case: a real worktree on disk reached
	// through a symlinked parent must still be detected via EvalSymlinks.
	root := t.TempDir()
	wt := filepath.Join(root, ".worktrees", "alpha", "src")
	if err := os.MkdirAll(wt, 0755); err != nil {
		t.Fatalf("mkdir wt: %v", err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}
	// cwd via the symlinked parent
	via := filepath.Join(link, ".worktrees", "alpha", "src")
	feat, _ := worktree.DetectFeatureFromCwd(via)
	if feat != "alpha" {
		t.Fatalf("symlinked cwd should resolve feature, got %q", feat)
	}
}

func TestIsInsideWorktree(t *testing.T) {
	in := filepath.Join(string(filepath.Separator), "repo", ".worktrees", "alpha")
	out := filepath.Join(string(filepath.Separator), "repo", "src")
	if !worktree.IsInsideWorktree(in) {
		t.Fatalf("expected inside-worktree=true for %q", in)
	}
	if worktree.IsInsideWorktree(out) {
		t.Fatalf("expected inside-worktree=false for %q", out)
	}
}

func TestExists_DirectoryAndMissing(t *testing.T) {
	repo := t.TempDir()
	if worktree.Exists(repo, "alpha") {
		t.Fatal("Exists must be false before creation")
	}
	if err := os.MkdirAll(filepath.Join(repo, ".worktrees", "alpha"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if !worktree.Exists(repo, "alpha") {
		t.Fatal("Exists must be true after creation")
	}
}

func TestExists_FilePathReturnsFalse(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".worktrees"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create `.worktrees/alpha` as a regular file, not a directory.
	if err := os.WriteFile(filepath.Join(repo, ".worktrees", "alpha"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if worktree.Exists(repo, "alpha") {
		t.Fatal("Exists must reject regular files masquerading as worktrees")
	}
}
