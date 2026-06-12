package roadmap

import "strings"

// BacklogPhaseName is the canonical name of the validate-exempt phase that
// holds deferred findings captured by `centinela roadmap defer`.
const BacklogPhaseName = "Backlog"

// isBacklogPhaseName reports whether a phase name denotes the Backlog phase.
// Mirrors isBootstrapPhaseName: lower-cased, trimmed, exact canonical match.
func isBacklogPhaseName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), BacklogPhaseName)
}

// IsBacklogPhaseName is the exported form for callers outside this package
// (e.g. the UI render layer skipping Backlog in the phase loop).
func IsBacklogPhaseName(name string) bool {
	return isBacklogPhaseName(name)
}

// IsBacklogFeature reports whether feature lives in a Backlog phase.
func IsBacklogFeature(r *Roadmap, feature string) bool {
	if r == nil {
		return false
	}
	for _, p := range r.Phases {
		if !isBacklogPhaseName(p.Name) {
			continue
		}
		for _, f := range p.Features {
			if f.Name == feature {
				return true
			}
		}
	}
	return false
}

// BacklogFeatures returns every Feature inside a Backlog phase, declared order.
func BacklogFeatures(r *Roadmap) []Feature {
	if r == nil {
		return nil
	}
	var out []Feature
	for _, p := range r.Phases {
		if !isBacklogPhaseName(p.Name) {
			continue
		}
		out = append(out, p.Features...)
	}
	return out
}

// NonBacklogFeatureSet returns the names of all features NOT in a Backlog
// phase. This is the single coverage set ValidateAnalysis/ValidateQuality and
// readiness use, so Backlog findings are exempt in exactly one place.
func NonBacklogFeatureSet(r *Roadmap) map[string]bool {
	out := map[string]bool{}
	if r == nil {
		return out
	}
	for _, p := range r.Phases {
		if isBacklogPhaseName(p.Name) {
			continue
		}
		for _, f := range p.Features {
			out[f.Name] = true
		}
	}
	return out
}
