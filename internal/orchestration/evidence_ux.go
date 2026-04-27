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

func normalizeUXTag(tag string) string {
	text := strings.ToLower(strings.TrimSpace(tag))
	text = strings.ReplaceAll(text, "_", "-")
	return strings.ReplaceAll(text, " ", "-")
}
