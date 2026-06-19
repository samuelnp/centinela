package synthesize

import (
	"sort"

	"github.com/samuelnp/centinela/internal/analyze"
)

// Inferer maps an Inventory to an archetype Inference. It is an interface so a
// future LLM-backed inferer can replace the deterministic default without
// touching cmd/ or the synthesizer.
type Inferer interface {
	Infer(analyze.Inventory) Inference
}

type ruleInferer struct{ rules []rule }

// NewInferer returns the default deterministic, rule-table-driven inferer.
func NewInferer() Inferer { return ruleInferer{rules: rules} }

// Infer scores every archetype against the inventory's signals, ranks them
// deterministically (total desc, then archetype name asc), and derives the
// winner's confidence and whether the top two were within the tie margin.
func (r ruleInferer) Infer(inv analyze.Inventory) Inference {
	s := newSignals(inv)
	byArch := map[Archetype]*Score{}
	for _, rl := range r.rules {
		if !rl.match(s) {
			continue
		}
		sc := byArch[rl.arch]
		if sc == nil {
			sc = &Score{Archetype: rl.arch}
			byArch[rl.arch] = sc
		}
		sc.Total += rl.weight
		sc.Signals = append(sc.Signals, Signal{Reason: rl.reason, Weight: rl.weight})
	}
	scores := rank(byArch)
	return classify(scores)
}

func rank(byArch map[Archetype]*Score) []Score {
	out := make([]Score, 0, len(byArch))
	for _, sc := range byArch {
		out = append(out, *sc)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Archetype < out[j].Archetype
	})
	return out
}

// classify derives Best/Confidence/Ambiguous from ranked scores. No score → a
// low-confidence Custom fallback; a top-two tie within margin 1 → ambiguous+low.
func classify(scores []Score) Inference {
	if len(scores) == 0 || scores[0].Total == 0 {
		return Inference{Best: Custom, Confidence: Low, Scores: scores}
	}
	top := scores[0].Total
	margin := top
	if len(scores) > 1 {
		margin = top - scores[1].Total
	}
	ambiguous := len(scores) > 1 && scores[1].Total > 0 && margin <= 1
	conf := Low
	switch {
	case ambiguous:
		conf = Low
	case top >= 6 && margin >= 3:
		conf = High
	case top >= 3:
		conf = Medium
	}
	return Inference{Best: scores[0].Archetype, Confidence: conf, Scores: scores, Ambiguous: ambiguous}
}
