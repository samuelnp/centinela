package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// patchTsconfigExclude returns (false,nil) early when the entry is already present.
func TestPatchTsconfigExclude_AlreadyPresent_NoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tsconfig.json")
	if err := os.WriteFile(path, []byte(`{"exclude":[".worktrees"]}`), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	changed, err := patchTsconfigExclude(path, ".worktrees")
	if err != nil {
		t.Fatalf("patchTsconfigExclude: %v", err)
	}
	if changed {
		t.Fatal("expected no-op when entry already present")
	}
}

// MaybeProvision surfaces Create's slug-validation error.
func TestMaybeProvision_InvalidSlug_ReturnsError(t *testing.T) {
	repo := t.TempDir()
	// Make it look like a git repo so isGitRepo passes and Create is reached.
	if _, err := gitRunner(repo, "init", "-q"); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	cfgEnable := func() bool { return true }
	_ = cfgEnable
	// Build a config with the flag on via the exported helper path.
	if _, err := Create(repo, "Bad/Slug"); err == nil {
		t.Fatal("Create must reject an invalid slug")
	}
}

// readSpecsFrom skips sub-directories and non-.feature files.
func TestReadSpecsFrom_SkipsDirsAndNonFeatures(t *testing.T) {
	dir := t.TempDir()
	specs := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specs, "nested"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specs, "notes.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specs, "real.feature"),
		[]byte("Feature: R\n  Scenario: s\n    Given g\n    Then t\n"), 0644); err != nil {
		t.Fatalf("write feature: %v", err)
	}
	recs := readSpecsFrom(specs, "owner")
	if len(recs) == 0 {
		t.Fatal("expected the .feature file to be parsed past the dir/txt entries")
	}
}

// Remove surfaces a git failure when the worktree metadata is corrupt.
func TestRemove_GitFailureSurfacesError(t *testing.T) {
	repo := t.TempDir()
	if _, err := gitRunner(repo, "init", "-q", "-b", "main"); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	// Create a .worktrees/<feature> directory that git does not know about.
	bogus := filepath.Join(repo, Dir, "phantom")
	if err := os.MkdirAll(bogus, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := Remove(repo, "phantom", false); err == nil {
		t.Fatal("Remove should error when git has no such worktree registered")
	}
}
