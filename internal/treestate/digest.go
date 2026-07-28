package treestate

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// excluded is the output directory of verification, never its input (D3a).
const excluded = ".workflow/"

// Digest reduces `git status --porcelain=v1` and `git diff HEAD` to a stable
// content hash, dropping every entry that only touches .workflow/. Status
// lines are sorted so the digest is order-independent.
func Digest(status, diff string) string {
	lines := filterStatus(status)
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n") + "\x00" + filterDiff(diff)))
	return fmt.Sprintf("sha256:%x", sum)
}

// filterStatus drops porcelain entries whose every path lies under .workflow/.
// A rename OUT of .workflow/ is a real change and is kept.
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

// filterDiff drops whole per-file sections of a unified diff whose paths lie
// under .workflow/. Splitting on the `diff --git` header keeps hunk bodies
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

func allExcluded(paths []string) bool {
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if !strings.HasPrefix(strings.Trim(strings.TrimSpace(p), `"`), excluded) {
			return false
		}
	}
	return true
}
