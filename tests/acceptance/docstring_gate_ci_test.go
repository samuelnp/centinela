package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Acceptance: specs/docstring-gate.feature
//
// The previous version of this test grepped validate.yml for `fetch-depth: 0`
// and called the gate "enforcing in CI". It was wrong: the failure was never
// depth, it was ref-name resolution, and a config-string assertion cannot see
// that. This drives the real binary in a reproduced actions/checkout ref state
// instead — full history, detached HEAD, no local branch ref for the base.

// setupCIShapeRepo builds an "upstream" repo, clones it, then detaches HEAD and
// deletes every local branch, leaving the base reachable only as origin/main.
func setupCIShapeRepo(t *testing.T, severity string) string {
	t.Helper()
	up := t.TempDir()
	runGit(t, up, "init", "-q", "-b", "main")
	runGit(t, up, "config", "user.email", "test@centinela.dev")
	runGit(t, up, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(up, "centinela.toml"), docstringToml(severity))
	mustWrite(t, filepath.Join(up, "src", "legacy.go"), "package a\n")
	runGit(t, up, "add", ".")
	runGit(t, up, "commit", "-q", "-m", "baseline")

	clone := filepath.Join(t.TempDir(), "clone")
	runGit(t, t.TempDir(), "clone", "-q", up, clone)
	runGit(t, clone, "config", "user.email", "test@centinela.dev")
	runGit(t, clone, "config", "user.name", "Test")
	runGit(t, clone, "checkout", "-q", "--detach", "HEAD")
	runGit(t, clone, "branch", "-D", "main")
	if _, err := os.Stat(filepath.Join(clone, ".git", "refs", "heads", "main")); err == nil {
		t.Fatal("fixture must leave no local main ref")
	}
	return clone
}

// Scenario: This repository's CI resolves a merge base so the gate enforces
func TestDG_CIRefShapeResolvesBaseAndEnforces(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupCIShapeRepo(t, "fail")

	mustWrite(t, filepath.Join(dir, "src", "undoc.go"), "package a\n\nfunc Undoc() {}\n")
	commit(t, dir, "add undocumented export")

	out, code := runCent(t, bin, dir, "docs", "lint", "--changed")
	if code != 1 {
		t.Fatalf("gate must ENFORCE in the CI ref shape, exit %d:\n%s", code, out)
	}
	mustContain(t, out, "since origin/main")
	mustContain(t, out, "src/undoc.go:3: func Undoc has no doc comment")
	mustNotContain(t, out, "scope unresolved")
	mustNotContain(t, out, "not found")
}

// The ratchet still holds in that shape: a legacy file left untouched on the
// base commit must not be dragged in by the origin/ fallback.
func TestDG_CIRefShapeStillRatchets(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupCIShapeRepo(t, "fail")
	mustWrite(t, filepath.Join(dir, "src", "new.go"), "package a\n\n// New is documented.\nfunc New() {}\n")
	commit(t, dir, "add documented export")

	out, code := runCent(t, bin, dir, "docs", "lint", "--changed")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, out)
	}
	mustContain(t, out, "since origin/main")
	mustContain(t, out, "All 1 exported identifiers")
}
