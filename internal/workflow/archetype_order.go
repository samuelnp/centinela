package workflow

// ArchetypeStepOrder returns the canonical-step subset/order for an archetype
// and whether the name is known. The returned slice is a fresh clone, so callers
// never alias DefaultStepOrder's backing array.
//
//	canonical : plan, code, tests, validate, docs  (= DefaultStepOrder)
//	hotfix    : code, tests, validate
//	refactor  : plan, code, tests, validate
//	spike     : plan, code  (no validate step = ungated by construction)
//
// spike omits the validate step on purpose: the ship gate keys on the presence
// of the "validate" step (complete.go), never on an archetype label, so spike is
// ungated by absence rather than by a bypass branch. A promoted spike is still
// re-validated step-agnostically at merge — see the safety argument in the plan.
func ArchetypeStepOrder(name string) ([]string, bool) {
	switch NormalizeArchetype(name) {
	case ArchetypeCanonical:
		return cloneOrder(DefaultStepOrder), true
	case ArchetypeHotfix:
		return []string{"code", "tests", "validate"}, true
	case ArchetypeRefactor:
		return []string{"plan", "code", "tests", "validate"}, true
	case ArchetypeSpike:
		return []string{"plan", "code"}, true
	default:
		return nil, false
	}
}
