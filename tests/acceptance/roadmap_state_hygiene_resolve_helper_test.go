// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// rshConflictRepo produces a REAL merge conflict: base is committed on main,
// theirs branches off and commits its own roadmap.json, main (ours) commits a
// diverging roadmap.json, then `git merge theirs` is attempted into main. The
// returned dir's working tree is left exactly as git's merge machinery leaves
// it — conflict markers on disk, roadmap.json unmerged in the index.
func rshConflictRepo(t *testing.T, base, ours, theirs string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "rsh@centinela.dev")
	runGit(t, dir, "config", "user.name", "RSH")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	rmPath := filepath.Join(dir, ".workflow", "roadmap.json")
	mustWrite(t, rmPath, base)
	commit(t, dir, "base")

	runGit(t, dir, "checkout", "-q", "-b", "theirs")
	mustWrite(t, rmPath, theirs)
	commit(t, dir, "theirs")

	runGit(t, dir, "checkout", "-q", "main")
	mustWrite(t, rmPath, ours)
	commit(t, dir, "ours")

	c := exec.Command("git", "merge", "--no-edit", "theirs")
	c.Dir = dir
	_, _ = c.CombinedOutput() // a conflicting merge exits non-zero — expected
	return dir
}

// rshUnmerged reports whether roadmap.json is unmerged in dir's index.
func rshUnmerged(t *testing.T, dir string) bool {
	t.Helper()
	return rshGitOut(t, dir, "ls-files", "--unmerged", "--", ".workflow/roadmap.json") != ""
}
