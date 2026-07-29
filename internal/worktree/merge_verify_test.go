package worktree

import (
	"errors"
	"strings"
	"testing"
)

// HEAD moved past `before` → the ref verifiably advanced.
func TestVerifyAdvance_RefAdvanced(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		return []byte("after-sha\n"), nil
	})
	o := &MergeOutcome{Branch: "feat"}
	if err := verifyAdvance(o, "/repo", "before-sha"); err != nil {
		t.Fatalf("verifyAdvance: %v", err)
	}
	if !o.RefAdvanced || o.AlreadyMerged {
		t.Fatalf("want RefAdvanced only, got %+v", o)
	}
}

// Unmoved HEAD + branch already an ancestor → honest AlreadyMerged.
func TestVerifyAdvance_UnmovedAncestor_AlreadyMerged(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("same-sha\n"), nil
		}
		return nil, nil // merge-base --is-ancestor succeeds
	})
	o := &MergeOutcome{Branch: "feat"}
	if err := verifyAdvance(o, "/repo", "same-sha"); err != nil {
		t.Fatalf("verifyAdvance: %v", err)
	}
	if o.RefAdvanced || !o.AlreadyMerged {
		t.Fatalf("want AlreadyMerged only, got %+v", o)
	}
}

// Neither advanced nor ancestor: git exited 0 without delivering — the
// exact false-success shape this feature exists to kill. Hard error.
func TestVerifyAdvance_Neither_HardError(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("same-sha\n"), nil
		}
		return nil, errors.New("exit status 1") // not an ancestor
	})
	o := &MergeOutcome{Branch: "feat"}
	err := verifyAdvance(o, "/repo", "same-sha")
	if err == nil || !strings.Contains(err.Error(), "did not advance") {
		t.Fatalf("want hard 'did not advance' error, got: %v", err)
	}
	if o.RefAdvanced || o.AlreadyMerged {
		t.Fatalf("no success flag may be set on the hard-error path: %+v", o)
	}
}

// A failing HEAD read is surfaced, never treated as "unmoved".
func TestVerifyAdvance_HeadReadError_Propagated(t *testing.T) {
	stubGit(t, func(string, ...string) ([]byte, error) {
		return []byte("fatal: bad revision"), errors.New("exit status 128")
	})
	o := &MergeOutcome{Branch: "feat"}
	err := verifyAdvance(o, "/repo", "before-sha")
	if err == nil || !strings.Contains(err.Error(), "cannot read HEAD") {
		t.Fatalf("want HEAD read error, got: %v", err)
	}
}
