package worktree

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// git merge exits 0, HEAD does not move, and the branch is not an ancestor:
// Merge must hard-error instead of returning a claimable outcome.
func TestMerge_GitZeroButRefUnmoved_HardError(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		switch args[0] {
		case "status":
			return []byte(""), nil
		case "symbolic-ref":
			return []byte("main\n"), nil
		case "rev-parse":
			return []byte("same-sha\n"), nil // before == after
		case "merge":
			return []byte("Already up to date."), nil
		case "merge-base":
			return nil, errors.New("exit status 1") // NOT an ancestor
		}
		return nil, nil
	})
	out, err := Merge("/repo", "feat", func(string) (bool, string) { return true, "" })
	if err == nil || !strings.Contains(err.Error(), "did not advance") {
		t.Fatalf("want 'did not advance' hard error, got: %v", err)
	}
	if out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("no success flag may survive the hard error: %+v", out)
	}
}

// Removal is only claimed when the directory is actually gone: a remove
// call that "succeeds" while the worktree dir survives fails the merge.
func TestMerge_WorktreeSurvivesRemoval_Errors(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, Dir, "feat"), 0o755); err != nil {
		t.Fatal(err)
	}
	revs := []string{"before-sha\n", "after-sha\n"}
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		switch args[0] {
		case "status":
			return []byte(""), nil
		case "symbolic-ref":
			return []byte("main\n"), nil
		case "rev-parse":
			out := revs[0]
			if len(revs) > 1 {
				revs = revs[1:]
			}
			return []byte(out), nil
		case "merge":
			return []byte("Merge made by the 'ort' strategy."), nil
		case "worktree":
			return []byte(""), nil // remove "succeeds" but the dir survives
		}
		return nil, nil
	})
	out, err := Merge(repo, "feat", func(string) (bool, string) { return true, "" })
	if err == nil || !strings.Contains(err.Error(), "still exists") {
		t.Fatalf("surviving worktree must fail the merge, got: %v", err)
	}
	if !out.RefAdvanced {
		t.Fatalf("the ref did advance before removal failed: %+v", out)
	}
}
