package orchestration

import (
	"fmt"
	"strings"
)

// Runner identifies the host that resolves the directive into concrete model
// IDs. The hook has no runtime runner signal at emit time, so it stays
// RunnerUnknown and emits a both-runner reference line.
type Runner string

const (
	RunnerClaude   Runner = "claude"
	RunnerOpenCode Runner = "opencode"
	RunnerUnknown  Runner = "unknown"
)

// tierModels maps each tier to its per-runner model IDs. A model refresh edits
// only this table.
var tierModels = map[Tier]map[Runner]string{
	TierReasoning: {RunnerClaude: "claude-opus-4-7", RunnerOpenCode: "anthropic/claude-opus-4-7"},
	TierBalanced:  {RunnerClaude: "claude-sonnet-4-6", RunnerOpenCode: "anthropic/claude-sonnet-4-6"},
	TierFast:      {RunnerClaude: "claude-haiku-4-5-20251001", RunnerOpenCode: "anthropic/claude-haiku-4-5"},
}

// ResolveModel resolves a role to its concrete model ID for the given runner.
// It applies the config override (models[roleSlug]) when present and valid,
// otherwise the role's default tier. On any missing mapping it returns the tier
// name with ok=false so the caller can warn; it never panics.
func ResolveModel(role Role, models map[string]string, runner Runner) (string, bool) {
	tier := DefaultTierForRole(role)
	if raw, ok := models[string(role)]; ok {
		if normalized, valid := NormalizeTier(raw); valid {
			tier = normalized
		}
	}
	byRunner, ok := tierModels[tier]
	if !ok {
		return string(tier), false
	}
	id, ok := byRunner[runner]
	if !ok {
		return string(tier), false
	}
	return id, true
}

// ModelReference renders ONE compact line listing each tier in play with BOTH
// runner IDs. Tiers are deduped and emitted in stable order so the orchestrator
// can pick the ID for its runner.
func ModelReference(tiers []Tier) string {
	seen := map[Tier]bool{}
	var parts []string
	for _, tier := range AllowedTiers() {
		if !containsTier(tiers, tier) || seen[tier] {
			continue
		}
		seen[tier] = true
		byRunner := tierModels[tier]
		parts = append(parts, fmt.Sprintf("%s: %s (claude) / %s (opencode)",
			tier, byRunner[RunnerClaude], byRunner[RunnerOpenCode]))
	}
	return strings.Join(parts, "; ")
}

func containsTier(tiers []Tier, target Tier) bool {
	for _, tier := range tiers {
		if tier == target {
			return true
		}
	}
	return false
}
