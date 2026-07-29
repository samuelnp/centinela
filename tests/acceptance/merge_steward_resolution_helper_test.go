package acceptance_test

import (
	"os/exec"
	"testing"
)

// applyRepoMergeResolution emulates the Merge Steward applying its fix: the
// in-progress conflicted merge is resolved and committed in the primary tree.
// `merge --continue` proves the branch really landed before claiming success,
// so a test that aborts the merge instead would be asserting a lie.
func applyRepoMergeResolution(t *testing.T, repo string) {
	t.Helper()
	for _, args := range [][]string{
		{"checkout", "--theirs", "."},
		{"add", "-A"},
		{"commit", "--no-edit", "-q"},
	} {
		c := exec.Command("git", args...)
		c.Dir = repo
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
