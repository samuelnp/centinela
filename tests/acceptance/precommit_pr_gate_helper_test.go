package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helpers for specs/precommit-and-pr-gate.feature acceptance tests. They build a
// real git repo with a centinela.toml enabling the G1 file-size fail gate, drive
// the built binary's precommit / precommit install|uninstall / pr-gate surfaces,
// and assert exit codes + marker/summary lines.

const pcToml = "[gates]\nfile_size = true\ni18n = false\n"

// pcLines returns an n-line Go source body (n > 100 violates the G1 fail gate).
func pcLines(n int) string {
	var b strings.Builder
	b.WriteString("package x\n")
	for i := 1; i < n; i++ {
		b.WriteString("// filler\n")
	}
	return b.String()
}

// pcGit runs git in dir, failing the test on error.
func pcGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// pcRepo creates a temp git repo on main with an initial commit and a
// centinela.toml enabling the G1 fail gate. extraToml is appended verbatim.
func pcRepo(t *testing.T, extraToml string) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	pcGit(t, dir, "init", "-q", "-b", "main")
	pcGit(t, dir, "config", "user.email", "qa@centinela.dev")
	pcGit(t, dir, "config", "user.name", "QA")
	writeFile(t, dir, "centinela.toml", pcToml+extraToml)
	writeFile(t, dir, "README.md", "seed\n")
	pcGit(t, dir, "add", ".")
	pcGit(t, dir, "commit", "-q", "-m", "seed")
	return dir
}

// hookPath is the conventional pre-commit hook location in a repo.
func hookPath(dir string) string {
	return filepath.Join(dir, ".git", "hooks", "pre-commit")
}

// pcBranch checks out a feature branch off main and commits one file so that
// `merge-base HEAD main` resolves to the seed commit and the changed-since-base
// set (the pr-gate input) contains exactly that file.
func pcBranch(t *testing.T, dir, rel, body string) {
	t.Helper()
	pcGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, rel, body)
	pcGit(t, dir, "add", rel)
	pcGit(t, dir, "commit", "-q", "-m", "change")
}
