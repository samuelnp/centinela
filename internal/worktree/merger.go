package worktree

import (
	"fmt"
	"strings"
)

// MergeOutcome captures the structured result of a merge attempt.
type MergeOutcome struct {
	Feature      string
	Branch       string
	TextConflict bool
	ValidateFail bool
	GitOutput    string
	ValidateOut  string
	WorktreeKept bool
	ConflictedPaths []string
}

// ValidateRunner runs `centinela validate` (or equivalent) against the merged
// tree at repo. It returns (passed, combined output).
type ValidateRunner func(repo string) (bool, string)

// Merge performs the hybrid merge sequence for a feature.
//
// Sequence:
//  1. Verify main is clean.
//  2. Attempt `git merge --no-ff <branch>`.
//  3. On clean text merge: run validate.
//  4. On either text conflict OR validate failure: stop and return the
//     outcome so the caller can invoke the Merge Steward.
//  5. On full success: remove the worktree.
func Merge(repo, feature string, run ValidateRunner) (MergeOutcome, error) {
	out := MergeOutcome{Feature: feature, Branch: branchName(feature)}
	if dirty, err := isDirty(repo); err != nil {
		return out, err
	} else if dirty {
		return out, fmt.Errorf("main working tree is dirty — commit or stash before merging %q", feature)
	}
	raw, err := gitRunner(repo, "merge", "--no-ff", out.Branch)
	out.GitOutput = strings.TrimSpace(string(raw))
	if err != nil {
		out.TextConflict = true
		out.ConflictedPaths = parseConflictedPaths(repo)
		out.WorktreeKept = true
		return out, nil
	}
	passed, vout := run(repo)
	out.ValidateOut = vout
	if !passed {
		out.ValidateFail = true
		out.WorktreeKept = true
		return out, nil
	}
	if err := Remove(repo, feature, false); err != nil {
		return out, err
	}
	return out, nil
}
