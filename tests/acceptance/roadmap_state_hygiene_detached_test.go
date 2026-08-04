// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os"
	"strings"
	"testing"
)

// Scenario: a mutation on a detached HEAD is not committed
//
// A commit made on a detached HEAD is reachable from nothing: the next checkout
// orphans it and the record it carried is destroyed. Before this was blocked,
// the command printed "✓ Committed" and the deferral was gone one checkout
// later — a loss path the auto-commit itself created.
func TestRsh_DetachedHeadIsNotCommitted(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	runGit(t, dir, "checkout", "-q", "--detach")
	before := rshCommitCount(t, dir)

	out, code := runCent(t, bin, dir, "roadmap", "defer", "detached-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("the mutation must still exit 0: %d\n%s", code, out)
	}
	containsAll(t, out, "not committed", "detached HEAD")
	if got := rshCommitCount(t, dir); got != before {
		t.Fatalf("no commit may be made on a detached HEAD: %d -> %d\n%s", before, got, out)
	}
	if _, err := os.Stat(dir + "/ROADMAP.md"); err != nil {
		t.Fatalf("ROADMAP.md must still be regenerated: %v", err)
	}
	if !strings.Contains(mustRead(t, dir+"/.workflow/roadmap.json"), "detached-thing") {
		t.Fatal("the finding must be on disk")
	}

	// The point of refusing: the record survives reattaching to a branch,
	// which is exactly what an orphaned commit would not have done.
	runGit(t, dir, "stash", "push", "-q", "-u")
	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "stash", "pop", "-q")
	if !strings.Contains(mustRead(t, dir+"/.workflow/roadmap.json"), "detached-thing") {
		t.Fatal("the deferral must survive a checkout back to a branch")
	}
}
