package acceptance_test

import (
	"path/filepath"
	"testing"
)

// Helpers for specs/docstring-gate.feature acceptance tests. Every fixture
// repo carries centinela.toml with [gates.docstring] already enabled in its
// baseline commit, so writing the config itself never counts as a "changed
// file" under the ratchet these tests exercise.

// docstringToml renders a [gates.docstring] block at the given severity,
// scoped to src/ so fixtures never depend on the real DefaultRoots list.
func docstringToml(severity string) string {
	return "[gates.docstring]\nenabled = true\nseverity = \"" + severity + "\"\nroots = [\"src\"]\n"
}

// setupDocstringRepo builds a git repo whose baseline ("main") commit
// carries centinela.toml with the gate enabled and an empty src/legacy.go
// (no exported identifiers, so it can never itself be a violation), then
// checks out "feature". Most scenarios add fixtures on top of this.
func setupDocstringRepo(t *testing.T, severity string) string {
	t.Helper()
	return setupDocstringRepoWithLegacy(t, severity, "package a\n")
}

// setupDocstringRepoWithLegacy is setupDocstringRepo with caller-chosen
// legacy.go contents, so a scenario can seed an undocumented legacy export
// that must never surface once the repo moves on to the feature branch.
func setupDocstringRepoWithLegacy(t *testing.T, severity, legacyBody string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@centinela.dev")
	runGit(t, dir, "config", "user.name", "Test")
	mustWrite(t, filepath.Join(dir, "centinela.toml"), docstringToml(severity))
	mustWrite(t, filepath.Join(dir, "src", "legacy.go"), legacyBody)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "baseline")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	return dir
}

// writeDocFile writes a Go fixture under src/ in dir without committing it,
// returning its repo-relative slash path.
func writeDocFile(t *testing.T, dir, rel, body string) string {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "src", rel), body)
	return "src/" + rel
}
