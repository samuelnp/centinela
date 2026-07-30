package orchestration

// The built-in tier -> runner -> model id table lives here so resolve.go keeps
// its resolution and rendering logic within the file-size budget.

// tierModels maps each tier to its per-runner built-in model IDs. A model
// refresh edits only this table — and an operator does not have to wait for a
// Centinela release: [orchestration.model_map].<tier>.<runner> in centinela.toml
// overrides any entry here (resolution precedence rule 2, ahead of these
// built-ins at rule 3).
//
// The claude column carries bare FAMILY ALIASES, never a dated snapshot: the
// Claude Code runner resolves an alias to the current family member, so the
// default can never name a retired model. The opencode column stays
// provider-qualified because a bare alias is not a valid opencode model
// reference. Every value here must also be a key in config.builtinModelCapability.
//
// The codex column is intentionally empty until codex-support lands; a missing
// codex entry falls through to precedence rule 4 (the tier name).
var tierModels = map[Tier]map[Runner]string{
	TierReasoning: {RunnerClaude: "opus", RunnerOpenCode: "anthropic/claude-opus-4-7"},
	TierBalanced:  {RunnerClaude: "sonnet", RunnerOpenCode: "anthropic/claude-sonnet-4-6"},
	TierFast:      {RunnerClaude: "haiku", RunnerOpenCode: "anthropic/claude-haiku-4-5"},
}

// TierModelIDs returns every built-in model id in stable tier/runner order.
// Exported so the config package can assert that each one classifies without
// importing orchestration's table as literals (config is the leaf package; the
// dependency runs config <- orchestration, never the other way).
func TierModelIDs() []string {
	var ids []string
	for _, tier := range AllowedTiers() {
		for _, runner := range referenceRunners {
			if id := tierModels[tier][runner]; id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
