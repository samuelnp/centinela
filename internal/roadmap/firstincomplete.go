package roadmap

// FirstIncomplete walks every phase in declared order and returns the first
// feature whose derived status is not "done". The second return value is false
// when the roadmap is nil/empty or every feature is already done (the caller
// should then present a "roadmap complete" state).
func FirstIncomplete(r *Roadmap) (string, bool) {
	if r == nil {
		return "", false
	}
	for _, phase := range r.Phases {
		for _, f := range phase.Features {
			if name, ok := FirstNotDone(f.Name); ok {
				return name, true
			}
		}
	}
	return "", false
}

// FirstNotDone is the shared per-feature predicate: it returns the feature name
// and true when the feature's derived status is not "done". It is reused by
// FirstIncomplete (all phases) and FirstIncompleteBootstrap (Phase 0 only).
func FirstNotDone(name string) (string, bool) {
	if FeatureStatus(name) != "done" {
		return name, true
	}
	return "", false
}
