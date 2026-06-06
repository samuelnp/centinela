package orchestration

import (
	"fmt"
	"strings"
)

// Runner identifies the host that resolves the directive into concrete model
// IDs. The hook has no runtime runner signal at emit time, so it stays
// RunnerUnknown and emits an all-runners reference line listing each runner's
// resolved model ID so the orchestrator can pick its own row.
type Runner string

const (
	RunnerClaude   Runner = "claude"
	RunnerOpenCode Runner = "opencode"
	RunnerCodex    Runner = "codex"
	RunnerUnknown  Runner = "unknown"
)

// referenceRunners are the runners enumerated, in stable order, on the
// all-runners reference line.
var referenceRunners = []Runner{RunnerClaude, RunnerOpenCode, RunnerCodex}

// tierModels maps each tier to its per-runner built-in model IDs. A model
// refresh edits only this table. The codex column is intentionally empty until
// codex-support lands; a missing codex entry falls through to precedence rule 4.
var tierModels = map[Tier]map[Runner]string{
	TierReasoning: {RunnerClaude: "claude-opus-4-7", RunnerOpenCode: "anthropic/claude-opus-4-7"},
	TierBalanced:  {RunnerClaude: "claude-sonnet-4-6", RunnerOpenCode: "anthropic/claude-sonnet-4-6"},
	TierFast:      {RunnerClaude: "claude-haiku-4-5-20251001", RunnerOpenCode: "anthropic/claude-haiku-4-5"},
}

// ModelReference renders ONE compact line listing each tier in play with every
// runner's built-in ID. Tiers are deduped and emitted in stable order so the
// orchestrator can pick the ID for its runner. A runner with no built-in entry
// for a tier renders the tier name (rule-4 fallback marker).
func ModelReference(tiers []Tier) string {
	seen := map[Tier]bool{}
	var parts []string
	for _, tier := range AllowedTiers() {
		if !containsTier(tiers, tier) || seen[tier] {
			continue
		}
		seen[tier] = true
		parts = append(parts, fmt.Sprintf("%s: %s", tier, tierReferenceColumns(tier)))
	}
	return strings.Join(parts, "; ")
}

// tierReferenceColumns renders "id (runner)" for every reference runner. A
// runner missing a built-in model for the tier renders the tier name so no
// other runner's concrete ID leaks under it.
func tierReferenceColumns(tier Tier) string {
	byRunner := tierModels[tier]
	cols := make([]string, 0, len(referenceRunners))
	for _, runner := range referenceRunners {
		id := byRunner[runner]
		if id == "" {
			id = string(tier)
		}
		cols = append(cols, fmt.Sprintf("%s (%s)", id, runner))
	}
	return strings.Join(cols, " / ")
}

func containsTier(tiers []Tier, target Tier) bool {
	for _, tier := range tiers {
		if tier == target {
			return true
		}
	}
	return false
}
