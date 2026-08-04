package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realDir is the canonical spelling of a directory. Tests must compare against
// it, not against the raw t.TempDir() string: on macOS that is a /var/folders
// path whose real location is /private/var/folders, and canonicalizing is
// exactly what the lock now does.
func realDir(t *testing.T, dir string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return dir
	}
	return resolved
}

// The lock must never become an untracked file in the working tree — dirtying
// the tree is the failure this whole feature exists to remove.
func TestStateLockPathIsNeverInTheWorkingTree(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := stateLockPath(filepath.Join(dir, RoadmapFile))
	if !strings.HasPrefix(got, filepath.Join(realDir(t, dir), ".git")+string(filepath.Separator)) {
		t.Fatalf("lock path %q must live inside the git directory", got)
	}
	if strings.Contains(got, string(filepath.Separator)+".workflow"+string(filepath.Separator)) {
		t.Fatalf("lock path %q must not be a sibling of roadmap.json", got)
	}
}

// A linked worktree carries `gitdir: <path>` instead of a .git directory.
func TestStateLockPathFollowsAWorktreeGitdirPointer(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "realgit")
	tree := filepath.Join(root, "tree")
	for _, d := range []string{real, filepath.Join(tree, ".workflow")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tree, ".git"), []byte("gitdir: "+real+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := stateLockPath(filepath.Join(tree, RoadmapFile))
	if filepath.Dir(got) != realDir(t, real) {
		t.Fatalf("lock path %q must resolve through the gitdir pointer to %q", got, real)
	}
}

// A relative gitdir pointer resolves against the worktree root, as git does.
func TestStateLockPathResolvesARelativeGitdirPointer(t *testing.T) {
	root := t.TempDir()
	tree := filepath.Join(root, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "sub", "gitdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, ".git"), []byte("gitdir: sub/gitdir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := stateLockPath(filepath.Join(tree, RoadmapFile)); filepath.Dir(got) != realDir(t, filepath.Join(tree, "sub", "gitdir")) {
		t.Fatalf("relative pointer not resolved: %q", got)
	}
}
