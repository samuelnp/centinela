// Package mcp exposes Centinela's governance as a versioned MCP server. It is an
// aggregator: it reuses internal/verdict.AssembleVerdict as the wire payload and
// imports the verdict/gates/verify/workflow/config engines plus the official Go
// MCP SDK, and is imported only by cmd/. Advisory-by-protocol — it returns a
// verdict (allow/warn/block) and never performs or blocks a write.
package mcp

import "github.com/samuelnp/centinela/internal/verdict"

// SchemaVersion is the wire-compat identifier harnesses pin against.
const SchemaVersion = "centinela.mcp/v1"

// nz coalesces a nil slice to an empty one so it marshals as a JSON array (`[]`)
// rather than `null` — the SDK validates tool output against the inferred schema,
// which requires arrays, and rejects null.
func nz[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// FeatureInput selects a feature; empty means the active feature.
type FeatureInput struct {
	Feature string `json:"feature" jsonschema:"feature slug to evaluate; empty uses the active feature"`
}

// VerifyInput selects a feature and step for claim verification.
type VerifyInput struct {
	Feature string `json:"feature" jsonschema:"feature slug; empty uses the active feature"`
}

// RulesInput takes no arguments.
type RulesInput struct{}

// GatesOutput is the run_gates result: gate lines plus the gates-scope decision.
type GatesOutput struct {
	Schema   string             `json:"schema"`
	Decision string             `json:"decision"`
	Gates    []verdict.GateLine `json:"gates"`
}

// VerifyOutput is the verify_claims result: check lines plus the verify-scope decision.
type VerifyOutput struct {
	Schema   string              `json:"schema"`
	Decision string              `json:"decision"`
	Checks   []verdict.CheckLine `json:"checks"`
}

// StateOutput is the workflow_state result: run provenance plus evidence index.
type StateOutput struct {
	Schema   string             `json:"schema"`
	Run      verdict.RunInfo    `json:"run"`
	Evidence []verdict.EvidLine `json:"evidence"`
}

// RulesOutput is the read_rules result: the governing rule surface.
type RulesOutput struct {
	Schema       string   `json:"schema"`
	Profile      string   `json:"profile"`
	Archetype    string   `json:"archetype"`
	MaxFileLines int      `json:"maxFileLines"`
	Gates        []string `json:"enabledGates"`
	Locales      []string `json:"locales"`
}
