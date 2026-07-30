package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// seedSpecRepo builds a hermetic main checkout with a committed "clash"
// scenario plus a companion-scenario file sharing one Given clause across two
// scenarios — the everyday shape that produced false positives (every idle
// worktree flagged, every companion scenario flagged) before the
// spec-conflict-false-positives hotfix.
func seedSpecRepo(t *testing.T, clashThen string) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "config", "user.email", "qa@centinela.dev")
	runGit(t, repo, "config", "user.name", "QA")
	mustWrite(t, filepath.Join(repo, ".gitignore"), ".worktrees/\n.workflow/\n")
	mustWrite(t, filepath.Join(repo, "centinela.toml"),
		"[validate]\ncommands = []\n[gates]\nfile_size = false\n")
	mustWrite(t, filepath.Join(repo, "specs", "login.feature"),
		"Feature: Login\n  Scenario: clash\n    Given user has account\n"+
			"    When user logs in\n    Then app routes to "+clashThen+"\n")
	mustWrite(t, filepath.Join(repo, "specs", "archetypes.feature"),
		"Feature: Archetypes\n"+
			"  Scenario: hotfix archetype resolves to code-tests-validate\n"+
			"    Given a feature started with the hotfix archetype\n"+
			"    Then the step list is code, tests, validate\n"+
			"  Scenario: active archetype is pinned\n"+
			"    Given a feature started with the hotfix archetype\n"+
			"    Then the archetype is recorded in the workflow state\n")
	commit(t, repo, "seed with specs")
	return repo
}

// addWorktreeBranch creates a real git worktree on a new branch from main's
// current HEAD, so it automatically carries byte-identical copies of every
// committed spec — exactly how a real feature worktree begins life.
func addWorktreeBranch(t *testing.T, repo, feature string) string {
	t.Helper()
	wt := filepath.Join(repo, ".worktrees", feature)
	runGit(t, repo, "worktree", "add", "-q", filepath.Join(".worktrees", feature), "-b", feature)
	return wt
}

// mainHeadSHA reads main's current commit so a blocked merge can be proven to
// have advanced nothing.
func mainHeadSHA(t *testing.T, repo string) string {
	t.Helper()
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = repo
	out, err := c.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}
