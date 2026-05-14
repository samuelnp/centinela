package worktree

import (
	"fmt"
	"regexp"
)

// featureSlugPattern enforces kebab-case ASCII slugs: lowercase letters and
// digits separated by single hyphens. Rejects shell metacharacters, path
// separators, and traversal sequences before they ever reach `git worktree
// add` or filesystem operations.
var featureSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateFeatureSlug returns nil when name is a safe kebab-case feature
// slug, or a descriptive error otherwise. Used to guard worktree paths and
// git branch names against injection and path-escape inputs (e.g.
// "alpha/../beta", "alpha;rm", "alpha beta").
func ValidateFeatureSlug(name string) error {
	if name == "" {
		return fmt.Errorf("worktree: feature name required")
	}
	if !featureSlugPattern.MatchString(name) {
		return fmt.Errorf("worktree: invalid feature slug %q: must match %s", name, featureSlugPattern.String())
	}
	return nil
}
