// Package synthesize is an aggregator that consumes the read-only
// internal/analyze Inventory contract to infer a project's best-fit architecture
// archetype and synthesize a draft PROJECT.md. It is deterministic (no LLM),
// byte-stable, and never overwrites an existing PROJECT.md. It imports only the
// internal/analyze domain package and the standard library; it is invoked from
// cmd/ and its types are rendered by internal/ui.
package synthesize

// Archetype is one of the supported architecture archetypes (architecture
// -overview.md), plus Custom for the inconclusive fallback.
type Archetype string

const (
	Hexagonal   Archetype = "hexagonal"
	RailsNative Archetype = "rails-native"
	NTier       Archetype = "n-tier"
	ECS         Archetype = "ecs"
	Modular     Archetype = "modular"
	Custom      Archetype = "custom"
)

// Confidence levels for an inference.
const (
	High   = "high"
	Medium = "medium"
	Low    = "low"
)

// Signal is one scored heuristic hit contributing to an archetype's total.
type Signal struct {
	Reason string
	Weight int
}

// Score is one archetype's accumulated weight and the signals that produced it.
type Score struct {
	Archetype Archetype
	Total     int
	Signals   []Signal
}

// Inference is the result of scoring all archetypes: the winner, its confidence,
// the ranked scores, and whether the top two were within the tie margin.
type Inference struct {
	Best       Archetype
	Confidence string
	Scores     []Score
	Ambiguous  bool
}

// Reasons returns the winning archetype's signal reasons, for rationale output.
func (inf Inference) Reasons() []string {
	for _, s := range inf.Scores {
		if s.Archetype == inf.Best {
			out := make([]string, 0, len(s.Signals))
			for _, sig := range s.Signals {
				out = append(out, sig.Reason)
			}
			return out
		}
	}
	return nil
}
