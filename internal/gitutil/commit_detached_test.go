package gitutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// F2: a commit on a detached HEAD is reachable from nothing — the next
// checkout orphans it and the record it carried is destroyed. Refuse instead.
func TestCommitPathsRefusesADetachedHead(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	mustGit(t, dir, "checkout", "-q", "--detach")
	if !DetachedHead(dir) {
		t.Fatal("a --detach checkout must report as detached")
	}
	if got := CommitBlockedReason(dir); got != "detached HEAD" {
		t.Fatalf("CommitBlockedReason = %q, want %q", got, "detached HEAD")
	}
	write(t, dir, ".workflow/roadmap.json", "{}\n")
	err := CommitPaths(dir, "chore(roadmap): defer detached-thing", []string{".workflow/roadmap.json"})
	if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("want a detached-HEAD refusal, got %v", err)
	}
	if got := strings.TrimSpace(mustGit(t, dir, "log", "-1", "--pretty=%s")); got != "seed" {
		t.Fatalf("no commit may be made on a detached HEAD, HEAD subject = %q", got)
	}
	mustGit(t, dir, "checkout", "-q", "main")
	if DetachedHead(dir) || CommitBlockedReason(dir) != "" {
		t.Fatal("reattaching must clear the block")
	}
}

// A rebase also detaches HEAD; the more specific operation must be reported.
func TestInProgressOperationOutranksDetachedHead(t *testing.T) {
	gitAvailable(t)
	dir := newRepo(t)
	mustGit(t, dir, "checkout", "-q", "--detach")
	gitDir := strings.TrimSpace(mustGit(t, dir, "rev-parse", "--absolute-git-dir"))
	if err := os.MkdirAll(filepath.Join(gitDir, "rebase-merge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := CommitBlockedReason(dir); got != "rebase in progress" {
		t.Fatalf("CommitBlockedReason = %q, want the rebase to outrank the detachment", got)
	}
}
