package worktree

import (
	"errors"
	"strings"
	"testing"
)

// Finding 2: a worktree registered outside `.worktrees/<feature>` used to pass
// verification trivially (os.Stat of a path that never existed), so the CLI
// claimed "removed its worktree" while it was still live and still listed.
func TestVerifyRemoved_StillRegisteredElsewhere_Refuses(t *testing.T) {
	porcelainStub(t, "worktree /repo\nbranch refs/heads/main\n\n"+
		"worktree /elsewhere/feat\nbranch refs/heads/feat\n\n")
	err := verifyRemoved(t.TempDir(), "feat")
	if err == nil || !strings.Contains(err.Error(), "still registered") {
		t.Fatalf("a still-registered worktree must refuse the claim, got: %v", err)
	}
	if !strings.Contains(err.Error(), "/elsewhere/feat") {
		t.Fatalf("refusal must name the surviving path: %v", err)
	}
}

func TestVerifyRemoved_RegistryUnreadable_Refuses(t *testing.T) {
	stubGit(t, func(string, ...string) ([]byte, error) {
		return nil, errors.New("exit status 128")
	})
	err := verifyRemoved(t.TempDir(), "feat")
	if err == nil || !strings.Contains(err.Error(), "cannot verify worktree removal") {
		t.Fatalf("an unreadable registry must refuse, not assume removal: %v", err)
	}
}

// verifyLanded is the `merge --continue` proof: ancestry in the target tree.
func TestVerifyLanded_NotAnAncestor_Refuses(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		return nil, errors.New("exit status 1") // merge-base says: not landed
	})
	o := MergeOutcome{Feature: "feat", Branch: "feat", TargetBranch: "main"}
	err := verifyLanded(&o, "/repo", "aaa")
	if err == nil || !strings.Contains(err.Error(), "was not completed") {
		t.Fatalf("a branch that did not land must refuse, got: %v", err)
	}
	if o.RefAdvanced || o.AlreadyMerged {
		t.Fatalf("no success flag may be set on refusal: %+v", o)
	}
}

func TestVerifyLanded_AdvancedAndAlreadyMerged(t *testing.T) {
	porcelainStub(t, "bbb\n") // is-ancestor exit 0, rev-parse HEAD = bbb
	o := MergeOutcome{Feature: "feat", Branch: "feat", TargetBranch: "main"}
	if err := verifyLanded(&o, "/repo", "aaa"); err != nil {
		t.Fatalf("landed branch must verify: %v", err)
	}
	if !o.RefAdvanced || o.AlreadyMerged {
		t.Fatalf("moved HEAD must read as advanced: %+v", o)
	}
	o2 := MergeOutcome{Feature: "feat", Branch: "feat", TargetBranch: "main"}
	if err := verifyLanded(&o2, "/repo", "bbb"); err != nil {
		t.Fatalf("unmoved HEAD with ancestry must verify: %v", err)
	}
	if o2.RefAdvanced || !o2.AlreadyMerged {
		t.Fatalf("unmoved HEAD must read as already merged: %+v", o2)
	}
}
