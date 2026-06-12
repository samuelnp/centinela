package roadmap

import (
	"fmt"
	"regexp"
	"strings"
)

// featureSlugPattern mirrors worktree.ValidateFeatureSlug — duplicated here so
// internal/roadmap keeps no import edge to internal/worktree (G2 import graph).
// Keep in sync with internal/worktree/slug.go.
var featureSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateSlug returns nil when slug is a safe kebab-case feature slug.
// mirrors worktree.ValidateFeatureSlug
func validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("feature slug required")
	}
	if !featureSlugPattern.MatchString(slug) {
		return fmt.Errorf("invalid feature slug %q: must match %s", slug, featureSlugPattern.String())
	}
	return nil
}

// validateNoCollision rejects a slug that already names any feature in any
// phase (Backlog or not), reporting which phase already holds it.
func validateNoCollision(slug string, existing map[string]string) error {
	if phase, ok := existing[slug]; ok {
		return fmt.Errorf("slug collision: %q already exists in phase %q", slug, phase)
	}
	return nil
}

// validateSummary rejects an empty or whitespace-only summary.
func validateSummary(summary string) error {
	if strings.TrimSpace(summary) == "" {
		return fmt.Errorf("summary is required and must not be empty")
	}
	return nil
}
