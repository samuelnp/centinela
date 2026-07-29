package worktree

import (
	"fmt"
	"os"
	"path/filepath"
)

// stewardEscalationDetail returns the steward report plus its proposed
// diff sibling (when present) so the caller can surface them to stderr.
func stewardEscalationDetail(repo, feature string) string {
	o := MergeOutcome{Feature: feature}
	detail := readFileOr(filepath.Join(repo, o.StewardHint()), "(no steward report found)")
	if diff, err := os.ReadFile(filepath.Join(repo, fmt.Sprintf(".workflow/%s-merge-steward.diff", feature))); err == nil {
		detail += "\n\n--- proposed diff ---\n" + string(diff)
	}
	return detail
}

func readFileOr(path, fallback string) string {
	if data, err := os.ReadFile(path); err == nil {
		return string(data)
	}
	return fallback
}
