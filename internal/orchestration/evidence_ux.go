package orchestration

import (
	"fmt"
	"strings"
)

var requiredUXTags = []string{
	"mobile-first",
	"visual-hierarchy",
	"typography-hierarchy",
	"responsive-layout",
	"loading-state",
	"empty-state",
	"error-state",
	"motion-and-reduced-motion",
}

func validateUXEvidence(path string, role Role, edgeCases []string, mobileFirst *bool) error {
	if role != RoleUXUISpecialist {
		return nil
	}
	if mobileFirst == nil || !*mobileFirst {
		return fmt.Errorf("ux-ui-specialist evidence must declare mobileFirst: true in: %s", path)
	}
	missing := missingUXTags(edgeCases)
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("ux-ui-specialist missing required ux edgeCases: %s in: %s", strings.Join(missing, ", "), path)
}

func missingUXTags(edgeCases []string) []string {
	tags := map[string]bool{}
	for _, edgeCase := range edgeCases {
		tags[normalizeUXTag(edgeCase)] = true
	}
	missing := []string{}
	for _, tag := range requiredUXTags {
		if !tags[tag] {
			missing = append(missing, tag)
		}
	}
	return missing
}

// normalizeUXTag reduces one edgeCase line to its tag so a descriptive entry
// ("mobile-first: renders at 80x24") satisfies the same requirement a bare
// token does — each UX fact is stated once, not twice. Order is part of the
// contract: trim+lowercase, cut at the FIRST colon, trim again (so
// "MOBILE_FIRST : Renders" works), then fold "_" and " " to "-".
//
// A degenerate entry (":", ": text", "", "   ") yields the empty string. That
// is inserted into the tag set like any other value and simply matches no
// required tag — no special-casing, no early continue, no panic.
func normalizeUXTag(tag string) string {
	text := strings.ToLower(strings.TrimSpace(tag))
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[:i]
	}
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "_", "-")
	return strings.ReplaceAll(text, " ", "-")
}
