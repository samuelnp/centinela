package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// annotateRoles builds the directive's role names (each annotated with its
// resolved model ID per runner), the required-evidence file list, and the
// deduped tiers in play for the model-reference line. The annotation stays
// runner-agnostic: it lists every runner's resolved ID so the orchestrator
// picks its own row. All resolution is delegated to internal/orchestration —
// cmd/ stays thin (G7).
func annotateRoles(feature string, roles []orchestration.Role, models orchestration.RoleModels, modelMap orchestration.ModelMap) (names, files []string, tiers []orchestration.Tier) {
	for _, role := range roles {
		names = append(names, fmt.Sprintf("%s (%s)", role, resolvedPerRunner(role, models, modelMap)))
		files = append(files, orchestration.MarkdownPath(feature, role))
		files = append(files, orchestration.JSONPath(feature, role))
		tiers = append(tiers, orchestration.RoleTier(role, models))
	}
	return names, files, tiers
}

// resolvedPerRunner formats "model: <id|tier> (<runner>)" for every runner so
// the directive is runner-agnostic.
func resolvedPerRunner(role orchestration.Role, models orchestration.RoleModels, modelMap orchestration.ModelMap) string {
	parts := make([]string, 0, len(orchestration.AllowedRunnerKeys()))
	for _, key := range orchestration.AllowedRunnerKeys() {
		id, _ := orchestration.ResolveModel(role, models, modelMap, orchestration.Runner(key))
		parts = append(parts, fmt.Sprintf("model: %s (%s)", id, key))
	}
	return strings.Join(parts, ", ")
}
