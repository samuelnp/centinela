package gitutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstLineReducesGitOutput(t *testing.T) {
	boom := errors.New("exit status 1")
	if got := firstLine("error: bad thing\nmore detail\n", boom); got != "error: bad thing" {
		t.Fatalf("firstLine = %q", got)
	}
	if got := firstLine("  \n\n", boom); got != "exit status 1" {
		t.Fatalf("a silent git must fall back to the exec error, got %q", got)
	}
}

// E1: a signing failure is the canonical "commit itself failed" case. The
// mutation must learn WHY, not just that something went wrong.
func TestCommitPathsReportsASigningFailure(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	mustGit(t, dir, "config", "commit.gpgsign", "true")
	mustGit(t, dir, "config", "gpg.program", "/nonexistent-gpg-binary")
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	err := CommitPaths(dir, "chore(roadmap): defer x", []string{".workflow/roadmap.json"})
	if err == nil || !strings.Contains(err.Error(), "commit failed") {
		t.Fatalf("want a commit-failed error, got %v", err)
	}
	if errors.Is(err, ErrNothingToCommit) {
		t.Fatal("a real failure must not masquerade as nothing-to-commit")
	}
}

// A path the mutation DELETED is still tracked, so it must stay in the
// pathspec — otherwise the removal would silently never be committed.
func TestCommitPathsCommitsADeletedTrackedPath(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	if err := CommitPaths(dir, "add", []string{".workflow/roadmap.json"}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, ".workflow/roadmap.json")); err != nil {
		t.Fatal(err)
	}
	if err := CommitPaths(dir, "remove", []string{".workflow/roadmap.json"}); err != nil {
		t.Fatalf("a deletion must be committable: %v", err)
	}
	if out := mustGit(t, dir, "show", "--name-status", "--pretty=format:", "HEAD"); !strings.HasPrefix(out, "D") {
		t.Fatalf("want a deletion in the commit, got %q", out)
	}
}

// A read-only git directory fails at `git add`, before any commit is attempted:
// the mutation must still learn the cause rather than a bare exit status.
func TestCommitPathsReportsAStagingFailure(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	gitDir := filepath.Join(dir, ".git")
	if err := os.Chmod(gitDir, 0o500); err != nil {
		t.Skip("cannot make the git directory read-only here")
	}
	t.Cleanup(func() { os.Chmod(gitDir, 0o755) }) //nolint:errcheck
	err := CommitPaths(dir, "m", []string{".workflow/roadmap.json"})
	if err == nil || !strings.Contains(err.Error(), "git add failed") {
		t.Fatalf("want a staging failure, got %v", err)
	}
}

// A project that gitignores a roadmap-state path must not have its whole
// mutation commit taken down by `git add` refusing that one entry.
func TestCommitPathsSkipsAnIgnoredUntrackedPath(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	write(t, dir, ".gitignore", "ROADMAP.md\n")
	mustGit(t, dir, "add", ".gitignore")
	mustGit(t, dir, "commit", "-qm", "ignore the markdown")
	write(t, dir, "ROADMAP.md", "# Roadmap\n")
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	if err := CommitPaths(dir, "chore(roadmap): defer x",
		[]string{".workflow/roadmap.json", "ROADMAP.md"}); err != nil {
		t.Fatalf("an ignored entry must be skipped, not fatal: %v", err)
	}
	if got := strings.TrimSpace(mustGit(t, dir, "show", "--name-only", "--pretty=format:", "HEAD")); got != ".workflow/roadmap.json" {
		t.Fatalf("commit changed %q", got)
	}
	if !isIgnored(dir, "ROADMAP.md") || isTracked(dir, "ROADMAP.md") {
		t.Fatal("the ignored path must stay untracked")
	}
	if isIgnored(dir, ".workflow/roadmap.json") {
		t.Fatal("a non-matching path must not report as ignored")
	}
}
