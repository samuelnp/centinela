package brownmap

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteDraft_RelativePathUsesCwd exercises the dir == "." / "" branch of
// atomicWrite by writing to a bare filename in a temp working directory.
func TestWriteDraft_RelativePathUsesCwd(t *testing.T) {
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	wrote, err := WriteDraft("draft.json", samplePlan())
	if err != nil || wrote != "draft.json" {
		t.Fatalf("relative WriteDraft err=%v wrote=%q", err, wrote)
	}
	if _, err := os.Stat("draft.json"); err != nil {
		t.Fatalf("relative draft not written: %v", err)
	}
}

// TestWriteDraft_RenameFailureWrapped forces os.Rename to fail (and the temp
// file to be cleaned up) by making the destination path an existing directory.
func TestWriteDraft_RenameFailureWrapped(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "draft.json")
	if err := os.Mkdir(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteDraft(dest, samplePlan()); err == nil {
		t.Fatal("rename over an existing directory must fail")
	}
	// the temp file must be cleaned up — only the dir remains in the parent.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temp file not cleaned up after rename failure: %v", entries)
	}
}
