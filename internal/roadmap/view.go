package roadmap

// BuildView projects a Roadmap into the deterministic RoadmapView emitted by
// `roadmap --json`. It is strictly read-only: it never mutates or persists the
// roadmap. Non-schedulable phases (Backlog/Baseline) are excluded to match
// Summary() and readiness scoping. Iteration is ordered-slice only so the
// marshalled bytes are stable across runs.
func BuildView(r *Roadmap) RoadmapView {
	view := RoadmapView{Phases: []PhaseView{}}
	if r == nil {
		return view
	}
	index := readinessIndex(r)
	for _, phase := range r.Phases {
		if isNonSchedulablePhase(phase.Name) {
			continue
		}
		pv := PhaseView{Name: phase.Name, Features: []FeatureView{}}
		for _, f := range phase.Features {
			fv := buildFeatureView(f, phase.Name, index[f.Name])
			pv.Features = append(pv.Features, fv)
			tally(&view.Counts, fv.Status)
		}
		view.Phases = append(view.Phases, pv)
	}
	return view
}

// readinessIndex derives readiness once and indexes it by feature name so
// BuildView reuses the existing classification instead of re-deriving deps.
func readinessIndex(r *Roadmap) map[string]FeatureReadiness {
	index := map[string]FeatureReadiness{}
	for _, fr := range DeriveReadiness(r) {
		index[fr.Name] = fr
	}
	return index
}

// buildFeatureView maps one Feature plus its readiness into a FeatureView.
// Readiness/BlockedBy are populated only for planned (ready|blocked) rows;
// DependsOn is always a non-nil slice so it serializes as [] rather than null.
func buildFeatureView(f Feature, phase string, fr FeatureReadiness) FeatureView {
	deps := f.DependsOn
	if deps == nil {
		deps = []string{}
	}
	fv := FeatureView{
		Name:      f.Name,
		Phase:     phase,
		Status:    FeatureStatus(f.Name),
		DependsOn: deps,
	}
	if fr.State == "ready" || fr.State == "blocked" {
		fv.Readiness = fr.State
		fv.BlockedBy = fr.BlockedBy
	}
	return fv
}

// tally increments the schedulable status counts for one feature status.
func tally(c *StatusCounts, status string) {
	switch status {
	case "done":
		c.Done++
	case "in-progress":
		c.InProgress++
	default:
		c.Planned++
	}
}
