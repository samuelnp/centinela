package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// symlinkRepo builds a repo at <root>/real and a symlink <root>/link -> real,
// returning both spellings of the SAME roadmap.json.
func symlinkRepo(t *testing.T) (realPath, linkPath string) {
	t.Helper()
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.MkdirAll(filepath.Join(real, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(real, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"Phase 1","features":[]}]}`
	if err := os.WriteFile(filepath.Join(real, RoadmapFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	return filepath.Join(real, RoadmapFile), filepath.Join(link, RoadmapFile)
}

// A symlinked spelling MUST hash to the same lock. When it did not, two lock
// files appeared side by side in one .git and the write race was silently back.
func TestStateLockPathIsSymlinkInvariant(t *testing.T) {
	realPath, linkPath := symlinkRepo(t)
	if got, want := stateLockPath(linkPath), stateLockPath(realPath); got != want {
		t.Fatalf("symlinked spelling produced a different lock:\n  via link: %s\n  via real: %s", got, want)
	}
}

// The everyday macOS shape: /tmp is itself a symlink to /private/tmp, so a repo
// under either spelling must resolve to one lock.
func TestStateLockPathIsInvariantAcrossTheSystemTempSymlink(t *testing.T) {
	const shadow, real = "/tmp", "/private/tmp"
	if resolved, err := filepath.EvalSymlinks(shadow); err != nil || resolved != real {
		t.Skipf("this platform does not symlink %s -> %s", shadow, real)
	}
	dir, err := os.MkdirTemp(shadow, "rsh-symlink-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) }) //nolint:errcheck
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	base := strings.TrimPrefix(dir, shadow)
	viaShadow := filepath.Join(shadow, base, RoadmapFile)
	viaReal := filepath.Join(real, base, RoadmapFile)
	if got, want := stateLockPath(viaShadow), stateLockPath(viaReal); got != want {
		t.Fatalf("/tmp vs /private/tmp produced different locks:\n  %s\n  %s", got, want)
	}
}
