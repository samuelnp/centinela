package worktree

import (
	"errors"
	"strings"
	"testing"
)

// Merge refuses a detached primary HEAD before touching anything: the
// validator must never run and no success flag may be set.
func TestMerge_DetachedPrimaryHead_Refused(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		switch args[0] {
		case "status":
			return []byte(""), nil // clean
		case "symbolic-ref":
			return nil, errors.New("exit status 1") // detached
		}
		t.Fatalf("unexpected git call %v past the detached-HEAD guard", args)
		return nil, nil
	})
	called := false
	out, err := Merge("/repo", "feat", func(string) (bool, string) { called = true; return true, "" })
	if err == nil || !strings.Contains(err.Error(), "detached HEAD") {
		t.Fatalf("want detached-HEAD refusal, got: %v", err)
	}
	if called || out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("nothing may run or be claimed after the refusal: called=%v %+v", called, out)
	}
}

// Merge refuses when the primary tree has the feature branch itself checked
// out: a self-merge no-ops and `isAncestor` is trivially true (a commit is
// its own ancestor), which would fabricate AlreadyMerged.
func TestMerge_FeatureBranchCheckedOutInPrimary_Refused(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		switch args[0] {
		case "status":
			return []byte(""), nil
		case "symbolic-ref":
			return []byte("feat\n"), nil // primary sits ON the feature branch
		}
		t.Fatalf("unexpected git call %v past the self-merge guard", args)
		return nil, nil
	})
	out, err := Merge("/repo", "feat", func(string) (bool, string) { return true, "" })
	if err == nil || !strings.Contains(err.Error(), "cannot merge a branch into itself") {
		t.Fatalf("want self-merge refusal, got: %v", err)
	}
	if out.RefAdvanced || out.AlreadyMerged {
		t.Fatalf("self-merge must not fabricate success: %+v", out)
	}
}
