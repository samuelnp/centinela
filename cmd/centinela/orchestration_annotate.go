package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// annotateRoles builds the directive's role names (each annotated with its
// resolved tier), the required-evidence file list, and the deduped tiers in
// play for the model-reference line. All resolution is delegated to
// internal/orchestration — cmd/ stays thin (G7).
func annotateRoles(feature string, roles []orchestration.Role, models map[string]string) (names, files []string, tiers []orchestration.Tier) {
	for _, role := range roles {
		tier := orchestration.DefaultTierForRole(role)
		if raw, ok := models[string(role)]; ok {
			if normalized, valid := orchestration.NormalizeTier(raw); valid {
				tier = normalized
			}
		}
		names = append(names, fmt.Sprintf("%s (model: %s)", role, tier))
		files = append(files, orchestration.MarkdownPath(feature, role))
		files = append(files, orchestration.JSONPath(feature, role))
		tiers = append(tiers, tier)
	}
	return names, files, tiers
}
