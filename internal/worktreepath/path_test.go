package worktreepath

import (
	"os"
	"path/filepath"
	"testing"
)

// The package is the shared `.worktrees/<feature>` vocabulary for
// internal/worktree AND internal/evidence, but had no colocated test: coverage
// is per-package, so exercising it through internal/worktree's delegation
// attributed 0% here. Every row asserts the ROOT too — the second return is
// what lets a caller derive from the worktree root instead of the CWD it
// happens to stand in, and a silently-wrong root is the E1 defect's shape.
func TestDetectFeatureTable(t *testing.T) {
	// A base that cannot exist keeps EvalSymlinks a no-op, so these rows test
	// the scan itself rather than the filesystem underneath it.
	const base = "/no-such-root-9f3/repo"
	cases := []struct {
		name, cwd, feature, root string
	}{
		{"worktree root", base + "/.worktrees/feat", "feat", base + "/.worktrees/feat"},
		{"one level down", base + "/.worktrees/feat/internal", "feat", base + "/.worktrees/feat"},
		{"deep subdirectory", base + "/.worktrees/feat/a/b/c", "feat", base + "/.worktrees/feat"},
		{"dot-dot climbs out", base + "/.worktrees/feat/..", "", ""},
		{"redundant separators", base + "/.worktrees//feat//a", "feat", base + "/.worktrees/feat"},
		{"trailing .worktrees names nothing", base + "/.worktrees", "", ""},
		{"no worktrees segment", base + "/internal/evidence", "", ""},
		{"prefix near-miss", base + "/.worktreesX/feat", "", ""},
		{"suffix near-miss", base + "/x.worktrees/feat", "", ""},
		{"nested resolves outermost", base + "/.worktrees/outer/.worktrees/inner", "outer", base + "/.worktrees/outer"},
		{"empty cwd is the process cwd", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cwd == "" {
				t.Chdir(t.TempDir()) // deterministic: a temp dir has no segment
			}
			f, r := DetectFeature(tc.cwd)
			if f != tc.feature {
				t.Fatalf("DetectFeature(%q) feature = %q, want %q", tc.cwd, f, tc.feature)
			}
			if r != filepath.FromSlash(tc.root) && !(tc.root == "" && r == "") {
				t.Fatalf("DetectFeature(%q) root = %q, want %q", tc.cwd, r, tc.root)
			}
		})
	}
}

// A relative cwd must resolve exactly like its absolute form: the caller is a
// CLI reading os.Getwd(), which is absolute, but callers in tests are not.
func TestDetectFeatureResolvesRelativePaths(t *testing.T) {
	base := realTempDir(t)
	sub := filepath.Join(base, ".worktrees", "relfeat", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(base, ".worktrees", "relfeat"))
	f, r := DetectFeature("pkg")
	if f != "relfeat" || r != filepath.Join(base, ".worktrees", "relfeat") {
		t.Fatalf("relative cwd resolved to %q/%q, want relfeat rooted at the checkout", f, r)
	}
}

// Symlinks are resolved BEFORE the scan (the documented /tmp -> /private/tmp
// case): a link pointing into a worktree must still name that worktree.
func TestDetectFeatureResolvesSymlinkIntoWorktree(t *testing.T) {
	base := realTempDir(t)
	target := filepath.Join(base, ".worktrees", "symfeat", "deep")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	f, r := DetectFeature(link)
	if f != "symfeat" || r != filepath.Join(base, ".worktrees", "symfeat") {
		t.Fatalf("symlinked cwd resolved to %q/%q, want symfeat rooted at the checkout", f, r)
	}
}

// realTempDir returns t.TempDir with symlinks resolved, so expectations built
// from it match what DetectFeature returns after its own EvalSymlinks.
func realTempDir(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	if r, err := filepath.EvalSymlinks(d); err == nil {
		return r
	}
	return d
}
