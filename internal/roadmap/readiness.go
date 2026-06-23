package roadmap

// FeatureReadiness is the derived readiness state for a single feature.
// States: "done", "in-progress", "ready", "blocked". Never persisted.
type FeatureReadiness struct {
	Name      string
	State     string
	BlockedBy []string // names of unmet deps; populated when State == "blocked"
}

// DeriveReadiness classifies every feature in declared order.
// A planned feature with all deps done is "ready"; with any dep not done it is
// "blocked". done/in-progress features pass through from FeatureStatus.
func DeriveReadiness(r *Roadmap) []FeatureReadiness {
	if r == nil {
		return nil
	}
	var result []FeatureReadiness
	for _, phase := range r.Phases {
		if isNonSchedulablePhase(phase.Name) {
			continue // Backlog findings + Baseline entries are never ready/blocked/startable
		}
		for _, f := range phase.Features {
			result = append(result, classifyFeature(f))
		}
	}
	return result
}

// ReadySet returns the names of all "ready" features in declared order.
func ReadySet(r *Roadmap) []string {
	var out []string
	for _, fr := range DeriveReadiness(r) {
		if fr.State == "ready" {
			out = append(out, fr.Name)
		}
	}
	return out
}

// UnmetDependencies returns the dep names whose status is not "done" for the
// named feature. Returns nil when all deps are done or the feature has none.
func UnmetDependencies(r *Roadmap, feature string) []string {
	if r == nil {
		return nil
	}
	for _, phase := range r.Phases {
		for _, f := range phase.Features {
			if f.Name != feature {
				continue
			}
			return collectUnmet(f.DependsOn)
		}
	}
	return nil
}

func classifyFeature(f Feature) FeatureReadiness {
	status := FeatureStatus(f.Name)
	switch status {
	case "done":
		return FeatureReadiness{Name: f.Name, State: "done"}
	case "in-progress":
		return FeatureReadiness{Name: f.Name, State: "in-progress"}
	default:
		unmet := collectUnmet(f.DependsOn)
		if len(unmet) == 0 {
			return FeatureReadiness{Name: f.Name, State: "ready"}
		}
		return FeatureReadiness{Name: f.Name, State: "blocked", BlockedBy: unmet}
	}
}

func collectUnmet(deps []string) []string {
	var unmet []string
	for _, dep := range deps {
		if FeatureStatus(dep) != "done" {
			unmet = append(unmet, dep)
		}
	}
	return unmet
}
