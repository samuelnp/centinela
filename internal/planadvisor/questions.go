package planadvisor

type question struct {
	Lens string
	Text string
	Ask  bool
}

func selectQuestions(c coverage, limit int, mode string) []question {
	force := mode == "always"
	all := []question{
		{"big-thinker", "What exact user or operator pain are we solving, and for whom?", force || !c.Problem},
		{"big-thinker", "What is explicitly in scope for the first version, and what should stay out?", force || !c.Scope},
		{"feature-specialist", "What observable behaviors or acceptance criteria must the spec guarantee?", force || !c.Acceptance},
		{"feature-specialist", "What should the primary mobile-first flow prioritize on small screens and touch devices?", c.UserFacing && (force || !c.MobileFirst)},
		{"feature-specialist", "Which edge cases or invalid states must the feature handle explicitly?", force || !c.EdgeCases},
		{"feature-specialist", "What should loading, empty, and error states communicate to the user?", c.UserFacing && (force || !c.UXStates)},
		{"big-thinker", "What constraints or non-negotiables should shape the design and rollout?", force || !c.Constraints},
		{"big-thinker", "What regressions, tradeoffs, or failure modes should we plan around now?", force || !c.Risks},
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
