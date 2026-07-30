package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitOutTruthful runs git in dir and returns trimmed stdout.
func gitOutTruthful(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// THE regression at the cmd layer: runMerge invoked with the CWD inside the
// feature worktree must merge in the PRIMARY tree (advancing main), remove
// the worktree, and print the one-line primary-tree notice.
func TestRunMerge_FromInsideWorktree_AdvancesMainAndPrintsNotice(t *testing.T) {
	d := seedCleanMergeRepo(t, "omega") // cwd = d, worktree omega committed
	before := gitOutTruthful(t, d, "rev-parse", "main")
	wt := filepath.Join(d, ".worktrees", "omega")
	if err := os.Chdir(wt); err != nil { // the old-bug CWD shape
		t.Fatal(err)
	}
	var err error
	out := captureStdout(t, func() { err = runMerge(nil, []string{"omega"}) })
	if err != nil {
		t.Fatalf("merge from worktree CWD must succeed: %v", err)
	}
	if !strings.Contains(out, "Merging in primary working tree") {
		t.Fatalf("want the primary-tree notice, got: %s", out)
	}
	if gitOutTruthful(t, d, "rev-parse", "main") == before {
		t.Fatal("main did NOT advance in the primary tree — the false-success bug")
	}
	gitOutTruthful(t, d, "merge-base", "--is-ancestor", "omega", "main")
	if _, e := os.Stat(wt); !os.IsNotExist(e) {
		t.Fatalf("worktree must be removed; stat err=%v", e)
	}
	if !strings.Contains(out, `Merged "omega" into main`) {
		t.Fatalf("want the delivery confirmation, got: %s", out)
	}
}

// An unresolvable primary tree (CWD outside any git repo) refuses before
// attempting anything — no merge, no success line.
func TestRunMerge_UnresolvablePrimaryTree_Refuses(t *testing.T) {
	chdir(t, t.TempDir())
	var err error
	out := captureStdout(t, func() { err = runMerge(nil, []string{"omega"}) })
	if err == nil || !strings.Contains(err.Error(), "cannot resolve primary working tree") {
		t.Fatalf("want the primary-tree refusal, got: %v", err)
	}
	if strings.Contains(out, "Merged") {
		t.Fatalf("no success may be printed on refusal: %s", out)
	}
}
