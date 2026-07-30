package worktree

import (
	"fmt"
	"strings"
)

// SpecConflict names the spec file and scenario that two in-flight worktrees
// define differently. OwnerA is always the feature being merged.
type SpecConflict struct {
	FeatureA string // spec file name as seen in the merging worktree
	OwnerA   string // feature slug that owns FeatureA (the merging feature)
	FeatureB string // spec file name as seen in the other worktree
	OwnerB   string // feature slug of the other in-flight worktree
	Scenario string
	Given    string
}

// maxFormattedConflicts caps FormatSpecConflicts output; the formatted string
// is embedded in a CLI error message, so an unbounded list is unreadable.
const maxFormattedConflicts = 10

// DetectSpecConflicts reports *parallel divergence*: a scenario that the
// merging feature's worktree and a DIFFERENT in-flight worktree both edited
// away from main, in the same spec file, into different content. That is the
// only case where merging one feature silently overwrites another's work.
//
// Deliberately NOT conflicts (each was a false positive before the
// spec-conflict-false-positives hotfix):
//   - main vs. the merging worktree — that is supersession, which the merge
//     exists to apply;
//   - a bystander worktree still carrying main's untouched copy;
//   - byte-identical copies of a spec carried by several worktrees;
//   - two companion scenarios inside one file sharing a Given clause;
//   - two different files sharing a Given clause.
func DetectSpecConflicts(repo, mergingFeature string) []SpecConflict {
	merging := readSpecsFrom(worktreeSpecsDir(repo, mergingFeature), mergingFeature)
	if len(merging) == 0 {
		return nil
	}
	baseline := mainSpecIndex(repo)
	var conflicts []SpecConflict
	seen := map[string]bool{}
	for _, other := range otherWorktreeSpecs(repo, mergingFeature) {
		conflicts = appendDivergences(conflicts, seen, merging, other, baseline)
	}
	return conflicts
}

// FormatSpecConflicts produces a human-readable, capped conflict list.
func FormatSpecConflicts(conflicts []SpecConflict) string {
	shown, suffix := conflicts, ""
	if len(shown) > maxFormattedConflicts {
		suffix = fmt.Sprintf("; … and %d more", len(shown)-maxFormattedConflicts)
		shown = shown[:maxFormattedConflicts]
	}
	lines := make([]string, 0, len(shown))
	for _, c := range shown {
		lines = append(lines, fmt.Sprintf("%s (%s) ↔ %s (%s) — scenario %q (Given: %s)",
			c.FeatureA, c.OwnerA, c.FeatureB, c.OwnerB, c.Scenario, c.Given))
	}
	return strings.Join(lines, "; ") + suffix
}
