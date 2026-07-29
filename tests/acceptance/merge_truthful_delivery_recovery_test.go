package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/merge-truthful-delivery.feature
// Scenario: a merge that lands but cannot remove the worktree reports both halves and stays recoverable
func TestAccMergeBusyWorktreeReportsBothHalvesAndRecovers(t *testing.T) {
	bin := buildCent(t)
	repo := mergeRepo(t, "high-score", false)
	wt := filepath.Join(repo, ".worktrees", "high-score")
	writeFile(t, wt, "scratch.txt", "untracked\n") // git refuses to remove this
	before := mtdaGit(t, repo, "rev-parse", "main")

	out, code := runCent(t, bin, repo, "merge", "high-score")
	if code == 0 {
		t.Fatalf("a worktree that survives removal must not exit 0:\n%s", out)
	}
	if mtdaGit(t, repo, "rev-parse", "main") == before {
		t.Fatalf("precondition: the merge itself should have landed:\n%s", out)
	}
	for _, want := range []string{"verified", "worktree removal failed", "--force-remove"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the half-success must state %q — an operator with an advanced "+
				"main and a bare error has no way forward:\n%s", want, out)
		}
	}
	// A plain re-run must repeat the same TRUTH, not regress into silence.
	rerun, code := runCent(t, bin, repo, "merge", "high-score")
	if code == 0 || !strings.Contains(rerun, "already merged into main — verified") {
		t.Fatalf("the re-run must say the merge already landed (code=%d):\n%s", code, rerun)
	}
	// And the named recovery command must actually work.
	rec, code := runCent(t, bin, repo, "merge", "high-score", "--force-remove")
	if code != 0 {
		t.Fatalf("--force-remove must complete the outstanding removal (code=%d):\n%s", code, rec)
	}
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be gone after the forced retry; err=%v\n%s", e, rec)
	}
}
