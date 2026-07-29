// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// tdGitRepoWithRealGitignore inits a git repo carrying the ACTUAL repo's
// .gitignore (not a hand-rolled pattern), so the test proves the real, edited
// .gitignore ignores the digest file — not just that git ignoring works.
func tdGitRepoWithRealGitignore(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitignore, err := os.ReadFile(filepath.Join(repoRoot(t), ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "qa@centinela.dev")
	runGit(t, dir, "config", "user.name", "QA")
	mustWrite(t, filepath.Join(dir, ".gitignore"), string(gitignore))
	tdWriteRoadmap(t, dir, tdRoadmap1)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "baseline")
	return dir
}

// Scenario: The digest state file never dirties the working tree
func TestTD_DigestStateFileNeverDirtiesWorkingTree(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdGitRepoWithRealGitignore(t)

	out, code := tdHookContext(t, bin, dir, tdSessionPayload("s-1"))
	if code != 0 {
		t.Fatalf("hook must exit 0: %s", out)
	}
	if _, err := os.Stat(tdDigestPath(dir)); err != nil {
		t.Fatalf("sanity: digest state should have been written: %v", err)
	}

	status := exec.Command("git", "status", "--porcelain")
	status.Dir = dir
	got, err := status.CombinedOutput()
	if err != nil {
		t.Fatalf("git status: %v\n%s", err, got)
	}
	if len(got) != 0 {
		t.Fatalf("git status --porcelain must report no changes, got:\n%s", got)
	}
}

// Scenario: Two worktrees keep independent digest state
func TestTD_TwoWorktreesKeepIndependentDigestState(t *testing.T) {
	bin := tdBuildBin(t)
	primary := tdGitRepoWithRealGitignore(t)
	wtRoot := t.TempDir()
	secondWT := filepath.Join(wtRoot, "second")
	runGit(t, primary, "worktree", "add", "-q", secondWT, "-b", "second")

	out1, code1 := tdHookContext(t, bin, primary, tdSessionPayload("s-1"))
	if code1 != 0 {
		t.Fatalf("first worktree hook must exit 0: %s", out1)
	}
	if _, err := os.Stat(tdDigestPath(primary)); err != nil {
		t.Fatalf("first worktree must have written its own digest state: %v", err)
	}

	out2, code2 := tdHookContext(t, bin, secondWT, tdSessionPayload("s-1"))
	if code2 != 0 {
		t.Fatalf("second worktree hook must exit 0: %s", out2)
	}
	mustContain(t, out2, tdRoadmapLine)
}
