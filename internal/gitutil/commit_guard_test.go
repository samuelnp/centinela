package gitutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsRepoAndHasHead(t *testing.T) {
	gitAvailable(t)
	plain := t.TempDir()
	if IsRepo(plain) || HasHead(plain) {
		t.Fatal("a plain directory is neither a repo nor has a HEAD")
	}
	empty := t.TempDir()
	mustGit(t, empty, "init", "-q", "-b", "main")
	if !IsRepo(empty) {
		t.Fatal("an initialized repo must report as a repo")
	}
	if HasHead(empty) {
		t.Fatal("a repo with no commits has no HEAD")
	}
	if reason := CommitBlockedReason(empty); reason != "no HEAD" {
		t.Fatalf("reason = %q, want %q", reason, "no HEAD")
	}
	dir := newRepo(t)
	if !HasHead(dir) || CommitBlockedReason(dir) != "" {
		t.Fatalf("a seeded repo must accept a commit: %q", CommitBlockedReason(dir))
	}
}

func TestCommitBlockedReasonNamesAMissingRepository(t *testing.T) {
	gitAvailable(t)
	if reason := CommitBlockedReason(t.TempDir()); reason != "no git repository" {
		t.Fatalf("reason = %q", reason)
	}
}

// A real conflicted merge: git itself refuses a partial commit here, so the
// guard must stop before attempting one.
func TestInProgressOperationDetectsARealMerge(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	mustGit(t, dir, "checkout", "-q", "-b", "other")
	write(t, dir, "seed.txt", "theirs\n")
	mustGit(t, dir, "commit", "-q", "-am", "theirs")
	mustGit(t, dir, "checkout", "-q", "main")
	write(t, dir, "seed.txt", "ours\n")
	mustGit(t, dir, "commit", "-q", "-am", "ours")
	if _, err := gitRun(dir, "merge", "other"); err == nil {
		t.Fatal("the merge was expected to conflict")
	}
	if got := InProgressOperation(dir); got != "merge in progress" {
		t.Fatalf("InProgressOperation = %q", got)
	}
	if got := CommitBlockedReason(dir); got != "merge in progress" {
		t.Fatalf("CommitBlockedReason = %q", got)
	}
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	err := CommitPaths(dir, "m", []string{".workflow/roadmap.json"})
	if err == nil || !strings.Contains(err.Error(), "merge in progress") {
		t.Fatalf("CommitPaths must refuse mid-merge, got %v", err)
	}
}

// A rebase leaves a directory, not a file, in the git dir — Stat must see both.
func TestInProgressOperationDetectsARebaseMarkerDirectory(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	gitDir := strings.TrimSpace(mustGit(t, dir, "rev-parse", "--absolute-git-dir"))
	if err := os.MkdirAll(filepath.Join(gitDir, "rebase-merge"), 0755); err != nil {
		t.Fatal(err)
	}
	if got := InProgressOperation(dir); got != "rebase in progress" {
		t.Fatalf("InProgressOperation = %q", got)
	}
}

func TestInProgressOperationIsEmptyOnACleanRepo(t *testing.T) {
	gitAvailable(t)
	if got := InProgressOperation(newRepo(t)); got != "" {
		t.Fatalf("clean repo reported %q", got)
	}
	if got := InProgressOperation(t.TempDir()); got != "" {
		t.Fatalf("non-repo reported %q", got)
	}
}
