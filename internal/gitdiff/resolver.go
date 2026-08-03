package gitdiff

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Summary describes the resolved diff context for the validate header.
type Summary struct {
	Base    string
	Files   int
	Degrade string // non-empty when diff resolution failed; e.g. "not a git repo"
}

// Resolver shells out to git. The runner field is overridable for tests.
type Resolver struct {
	Run func(name string, args ...string) (string, error)
}

// Default is a Resolver bound to the system git binary.
var Default = &Resolver{Run: runGit}

// ChangedFiles returns the union of tracked changes since merge-base and
// untracked files (when includeUntracked is true). The base is resolved by
// mergeBase, which retries the origin/<base> form so a CI checkout without a
// local branch ref still produces a real scope. On any git failure it returns
// (nil, summary with non-empty Degrade, nil error) so callers can degrade to
// full scan with a notice.
func (r *Resolver) ChangedFiles(base string, includeUntracked bool) (*Set, Summary, error) {
	if base == "" {
		base = "main"
	}
	summary := Summary{Base: base}

	resolved, mb, err := r.mergeBase(base)
	if err != nil {
		summary.Degrade = degradeReason(err, base)
		return nil, summary, nil
	}
	// Report the ref that actually resolved, not the one that was configured:
	// the validate header names the base, and naming an unused ref would be a
	// small lie about what the run compared against.
	summary.Base = resolved

	tracked, err := r.Run("git", "diff", "--name-only", "--diff-filter=ACMR", mb)
	if err != nil {
		summary.Degrade = fmt.Sprintf("git diff failed: %s", err)
		return nil, summary, nil
	}

	paths := splitNonEmpty(tracked)

	if includeUntracked {
		untracked, err := r.Run("git", "ls-files", "--others", "--exclude-standard")
		if err != nil {
			summary.Degrade = fmt.Sprintf("git ls-files failed: %s", err)
			return nil, summary, nil
		}
		paths = append(paths, splitNonEmpty(untracked)...)
	}

	set := NewSet(paths)
	summary.Files = set.Len()
	return set, summary, nil
}

func splitNonEmpty(s string) []string {
	out := make([]string, 0)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func runGit(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}
