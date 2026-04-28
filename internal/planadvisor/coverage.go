package planadvisor

import "strings"

type coverage struct {
	UserFacing, Problem, Scope, Constraints, Risks bool
	Acceptance, EdgeCases, MobileFirst, UXStates   bool
}

func scan(feature string) coverage {
	return scanTexts(loadArtifacts(feature))
}

func scanTexts(a artifacts) coverage {
	all := normalize(a.Brief + "\n" + a.Plan + "\n" + a.Spec + "\n" + a.Edge)
	return coverage{
		UserFacing:  has(a.Brief, "surface: user-facing"),
		Problem:     has(all, "## problem", " pain ", "who is the user", "who is affected"),
		Scope:       has(all, "## scope", "out of scope", "in scope", "## user stories"),
		Constraints: has(all, "## constraints", "non-goals", "latency", "compliance", "security"),
		Risks:       has(all, "## risks", "tradeoff", "trade-off", "regression", "failure mode"),
		Acceptance:  has(all, "## acceptance criteria", "scenario:", "given ", "when ", "then "),
		EdgeCases:   has(all, "## edge cases", "edge cases", "invalid input", "concurrency"),
		MobileFirst: has(all, "mobile-first", "mobile first"),
		UXStates:    has(all, "loading-state", "loading state") && has(all, "empty-state", "empty state") && has(all, "error-state", "error state"),
	}
}

func has(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(normalize(text), normalize(needle)) {
			return true
		}
	}
	return false
}

func normalize(text string) string {
	return strings.ToLower(strings.ReplaceAll(text, "_", "-"))
}
