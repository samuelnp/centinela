package worktree

import (
	"errors"
	"strings"
	"testing"
)

// Ancestry proven but HEAD unreadable: verifyLanded must surface the read
// failure rather than guess a wording for a delivery it cannot describe.
func TestVerifyLanded_HeadUnreadable_Errors(t *testing.T) {
	stubGit(t, func(_ string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "rev-parse" {
			return []byte("fatal: bad revision"), errors.New("exit status 128")
		}
		return nil, nil // merge-base --is-ancestor succeeds
	})
	o := MergeOutcome{Feature: "feat", Branch: "feat", TargetBranch: "main"}
	err := verifyLanded(&o, "/repo", "aaa")
	if err == nil || !strings.Contains(err.Error(), "cannot read HEAD") {
		t.Fatalf("an unreadable HEAD must surface, got: %v", err)
	}
	if o.RefAdvanced || o.AlreadyMerged {
		t.Fatalf("no flag may be set when HEAD could not be read: %+v", o)
	}
}
