package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mtdcSuccess = "and removed its worktree"

// mtdcStall drives a REAL text conflict from the given CWD and returns the
// primary tree, the worktree path and the target ref before anything moved.
// It asserts the stall itself: non-zero exit, no success line, and a pending
// marker in the PRIMARY tree (where the conflict actually lives).
func mtdcStall(t *testing.T, bin, feature string, fromWorktree bool) (repo, wt, before string) {
	t.Helper()
	repo = mergeRepo(t, feature, true)
	wt = filepath.Join(repo, ".worktrees", feature)
	before = mtdaGit(t, repo, "rev-parse", "main")
	cwd := repo
	if fromWorktree {
		cwd = wt
	}
	out, code := runCent(t, bin, cwd, "merge", feature)
	if code == 0 {
		t.Fatalf("a text conflict must stall, not succeed:\n%s", out)
	}
	if strings.Contains(out, mtdcSuccess) {
		t.Fatalf("a stalled merge must claim nothing:\n%s", out)
	}
	marker := filepath.Join(repo, ".workflow", feature+"-merge-pending.json")
	if _, e := os.Stat(marker); e != nil {
		t.Fatalf("the pending marker must be discoverable in the primary tree "+
			"whatever the invoking CWD (err=%v):\n%s", e, out)
	}
	return repo, wt, before
}

// mtdcAssertDelivered proves the delivery on the ref and on git's registry —
// never on the absence of errors.
func mtdcAssertDelivered(t *testing.T, repo, wt, feature, out string) {
	t.Helper()
	mtdaGit(t, repo, "merge-base", "--is-ancestor", feature, "main")
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be gone; err=%v\n%s", e, out)
	}
	if reg := mtdaGit(t, repo, "worktree", "list", "--porcelain"); strings.Contains(reg, wt) {
		t.Fatalf("worktree must be gone from the registry:\n%s", reg)
	}
	if !strings.Contains(out, mtdcSuccess) {
		t.Fatalf("a verified delivery must report success:\n%s", out)
	}
}
