package integration_test

import (
	"os/exec"
	"testing"

	"github.com/samuelnp/centinela/internal/gitutil"
)

func gitInit(t *testing.T, dir string, args ...[]string) {
	t.Helper()
	all := append([][]string{{"init"}}, args...)
	for _, a := range all {
		c := exec.Command("git", a...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", a, out)
		}
	}
}

// TestHasOriginRemoteRealRepos exercises the real `git` path: a fresh repo has
// no origin; after adding one, detection flips, and the delivery matrix follows.
func TestHasOriginRemoteRealRepos(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)

	ok, err := gitutil.HasOriginRemote(dir)
	if err != nil || ok {
		t.Fatalf("fresh repo should have no origin: ok=%v err=%v", ok, err)
	}
	// With no origin and worktree mode, only local merge is offered.
	if opts := gitutil.DeliveryOptions(ok, true); len(opts) != 1 || opts[0] != gitutil.OptionMerge {
		t.Fatalf("no-origin worktree should offer merge only, got %v", opts)
	}

	gitInit(t, dir, []string{"remote", "add", "origin", "https://example.com/x.git"})
	ok2, err := gitutil.HasOriginRemote(dir)
	if err != nil || !ok2 {
		t.Fatalf("after adding origin, detection should be true: ok=%v err=%v", ok2, err)
	}
	if opts := gitutil.DeliveryOptions(ok2, true); len(opts) != 2 {
		t.Fatalf("origin+worktree should offer both, got %v", opts)
	}
}
