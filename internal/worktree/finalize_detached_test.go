package worktree_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// A detached primary tree has no target branch to name, so `merge --continue`
// refuses before it can print a success line about "main".
func TestResolveMerge_DetachedPrimaryHead_Refuses(t *testing.T) {
	repo := resolveRepo(t, "iota")
	writeMarker(t, repo, "iota")
	landBranch(t, repo, "iota")
	gitOut(t, repo, "checkout", "-q", "--detach")
	res, err := worktree.ResolveMerge(repo, "iota", okValidator("complete"))
	if err == nil || !contains(err.Error(), "detached HEAD") {
		t.Fatalf("a detached primary tree must be refused, got: %v", err)
	}
	if res.Finalized {
		t.Fatal("must not finalize against a detached HEAD")
	}
}
