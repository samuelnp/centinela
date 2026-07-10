package worktree

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// currentBranch trims the symbolic-ref output down to the branch name.
func TestCurrentBranch_OnBranch(t *testing.T) {
	porcelainStub(t, "main\n")
	got, err := currentBranch("/repo")
	if err != nil {
		t.Fatalf("currentBranch: %v", err)
	}
	if got != "main" {
		t.Fatalf("currentBranch = %q, want main", got)
	}
}

// A detached HEAD (symbolic-ref failure) is a refusal with guidance.
func TestCurrentBranch_DetachedHead_Refused(t *testing.T) {
	stubGit(t, func(string, ...string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	})
	_, err := currentBranch("/repo")
	if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("detached HEAD must be refused, got: %v", err)
	}
}

// isAncestor maps merge-base exit status onto a boolean.
func TestIsAncestor_TrueAndFalse(t *testing.T) {
	porcelainStub(t, "")
	if !isAncestor("/repo", "feat") {
		t.Fatal("exit 0 must mean ancestor")
	}
	stubGit(t, func(string, ...string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	})
	if isAncestor("/repo", "feat") {
		t.Fatal("non-zero exit must mean not an ancestor")
	}
}

// verifyRemoved errors while the worktree directory survives and clears
// once it is actually gone — removal is a disk fact, not a git exit code.
func TestVerifyRemoved_DirSurvivesThenGone(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, Dir, "feat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	err := verifyRemoved(repo, "feat")
	if err == nil || !strings.Contains(err.Error(), "still exists") {
		t.Fatalf("surviving worktree dir must fail verification, got: %v", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := verifyRemoved(repo, "feat"); err != nil {
		t.Fatalf("gone dir must verify clean: %v", err)
	}
}
