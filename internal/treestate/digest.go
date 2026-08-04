package treestate

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/samuelnp/centinela/internal/roadmapstate"
)

// Digest reduces `git status --porcelain=v1`, `git diff HEAD` and the
// per-file hashes of untracked content to one stable value, dropping every
// entry whose paths are ALL roadmap state — verification's own output
// directory plus the generated ROADMAP.md (D3a, extended symmetrically by this
// feature: a mutation left uncommitted by disable_auto_commit must not stale a
// stamp either). Status lines are sorted so the digest is order-independent.
// This function stays pure: untracked hashes are computed by the caller (see
// HashUntracked) and passed in.
func Digest(status, diff string, untracked []string) string {
	lines := filterStatus(status)
	sort.Strings(lines)
	payload := strings.Join(lines, "\n") + "\x00" + filterDiff(diff) +
		"\x00" + strings.Join(untracked, "\n")
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(payload)))
}

// filterStatus drops porcelain entries whose every path is roadmap state.
// A rename OUT of roadmap state is a real change and is kept.
func filterStatus(status string) []string {
	kept := []string{}
	for _, ln := range strings.Split(status, "\n") {
		if len(ln) < 4 {
			continue
		}
		if !allExcluded(strings.Split(ln[3:], " -> ")) {
			kept = append(kept, ln)
		}
	}
	return kept
}

// filterDiff drops whole per-file sections of a unified diff whose paths are
// all roadmap state. Splitting on the `diff --git` header keeps hunk bodies
// attached to the file they belong to; sections are re-joined canonically so
// dropping a section cannot shift the surviving ones' separators.
func filterDiff(diff string) string {
	kept := []string{}
	for _, section := range strings.Split(diff, "\ndiff --git ") {
		header := strings.TrimPrefix(strings.SplitN(section, "\n", 2)[0], "diff --git ")
		if !allExcluded(stripPrefixes(strings.Fields(header))) {
			kept = append(kept, strings.TrimRight(section, "\n"))
		}
	}
	return strings.Join(kept, "\n")
}

// stripPrefixes removes git's a/ and b/ diff path prefixes.
func stripPrefixes(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimPrefix(strings.TrimPrefix(p, "a/"), "b/")
		out = append(out, strings.Trim(p, `"`))
	}
	return out
}

// allExcluded delegates to THE single definition of roadmap state, so the
// digest exclusion, the mutation pathspec and the freshness revision-range
// exemption can never drift apart. Covers is a strict subset test: a status
// entry mixing roadmap state with a source path is kept.
func allExcluded(paths []string) bool {
	return roadmapstate.Covers(paths)
}
