package planadvisor

import "strings"

type question struct {
	Lens string
	Text string
	Ask  bool
}

func selectQuestions(b bundle, limit int, mode string) []question {
	c := b.Coverage
	force := mode == "always"
	deps := joined(b.Dependencies)
	sibs := joined(b.Siblings)
	lesson := first(b.Lessons)
	quality := first(b.QualityNotes)
	all := []question{
		{"big-thinker", "What exact user or operator pain are we solving, and for whom?", force || !c.Problem},
		{"big-thinker", "What is explicitly in scope for the first version, and what should stay out?", force || !c.Scope},
		{"big-thinker", "Given dependencies " + deps + ", what sequencing, compatibility, or shared-contract constraints must we honor?", deps != "" && (force || !c.Constraints)},
		{"feature-specialist", "Roadmap quality notes include " + quality + ". What exact acceptance criteria should close that clarity gap?", quality != "" && (force || !c.Acceptance)},
		{"feature-specialist", "What observable behaviors or acceptance criteria must the spec guarantee?", force || !c.Acceptance},
		{"feature-specialist", "What should the primary mobile-first flow prioritize on small screens and touch devices?", c.UserFacing && (force || !c.MobileFirst)},
		{"feature-specialist", "Related edge-case lessons include " + lesson + ". Which of those failure patterns also apply here?", lesson != "" && (force || !c.EdgeCases)},
		{"big-thinker", "How should this feature stay consistent with same-phase siblings " + sibs + " without duplicating scope?", deps == "" && sibs != "" && (force || !c.Scope)},
		{"feature-specialist", "Which edge cases or invalid states must the feature handle explicitly?", force || !c.EdgeCases},
		{"feature-specialist", "What should loading, empty, and error states communicate to the user?", c.UserFacing && (force || !c.UXStates)},
		{"big-thinker", "What constraints or non-negotiables should shape the design and rollout?", force || !c.Constraints},
		{"big-thinker", "What regressions, tradeoffs, or failure modes should we plan around now?", force || !c.Risks},
		{"feature-specialist", "The ledger shows recurring gate failures (worst: " + worstGate(b) + "). What plan choices prevent that gate from biting this feature?", topFailureCount(b) >= 2},
	}
	out := []question{}
	for _, q := range all {
		if q.Ask {
			out = append(out, q)
		}
		if len(out) == limit {
			return out
		}
	}
	return out
}

func worstGate(b bundle) string {
	if len(b.Failures) == 0 {
		return ""
	}
	return b.Failures[0].Key
}

func topFailureCount(b bundle) int {
	if len(b.Failures) == 0 {
		return 0
	}
	return b.Failures[0].Count
}

func joined(items []string) string { return first([]string{strings.Join(items, ", ")}) }
func first(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}
