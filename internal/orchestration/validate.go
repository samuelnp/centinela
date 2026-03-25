package orchestration

import (
	"fmt"
	"os"
	"strings"
)

func ValidateStep(feature, step string) error {
	roles := RequiredRoles(step)
	if len(roles) == 0 {
		return nil
	}
	missing := []string{}
	invalid := []string{}
	for _, role := range roles {
		md := MarkdownPath(feature, role)
		if _, err := os.Stat(md); err != nil {
			missing = append(missing, md)
		}
		js := JSONPath(feature, role)
		if err := ValidateEvidence(js, feature, step, role); err != nil {
			if strings.Contains(err.Error(), "missing evidence") {
				missing = append(missing, js)
			} else {
				invalid = append(invalid, err.Error())
			}
		}
	}
	if len(missing) == 0 && len(invalid) == 0 {
		return nil
	}
	parts := []string{}
	if len(missing) > 0 {
		parts = append(parts, "missing: "+strings.Join(missing, ", "))
	}
	if len(invalid) > 0 {
		parts = append(parts, "invalid: "+strings.Join(invalid, "; "))
	}
	return fmt.Errorf("strict orchestration evidence failed for %q: %s", step, strings.Join(parts, " | "))
}
