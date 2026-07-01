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

// NonBacklogFeatureSet returns the names of all features in a schedulable
// phase. This is the single coverage set ValidateAnalysis/ValidateQuality and
// readiness use, so non-schedulable findings are exempt in exactly one place.
// It skips every non-schedulable phase (Backlog deferred findings and Baseline
// already-built capability) via the shared isNonSchedulablePhase predicate.
func NonBacklogFeatureSet(r *Roadmap) map[string]bool {
	// THE single draft exemption: an unscored draft in a schedulable phase carries
	// no analysis/quality entry until finalized, so the coverage set omits it. This
	// is the ONLY place the draft dimension changes the coverage set; the other
	// three draft readers (classifyFeature, Summary, BuildView) read f.Draft
	// directly, and dependency validation uses the fuller dependencyTargetSet.
	return schedulableFeatureSet(r, false)
}

// dependencyTargetSet returns every schedulable feature name that a dependsOn
// may legally reference — drafts INCLUDED, because a draft is a real feature you
// can depend on even though it is exempt from the analysis/quality coverage set.
func dependencyTargetSet(r *Roadmap) map[string]bool {
	return schedulableFeatureSet(r, true)
}

// schedulableFeatureSet collects feature names in schedulable (non-Backlog,
// non-Baseline) phases. When includeDrafts is false, drafts are omitted.
func schedulableFeatureSet(r *Roadmap, includeDrafts bool) map[string]bool {
	out := map[string]bool{}
	if r == nil {
		return out
	}
	for _, p := range r.Phases {
		if isNonSchedulablePhase(p.Name) {
			continue
		}
		for _, f := range p.Features {
			if f.Draft && !includeDrafts {
				continue
			}
			out[f.Name] = true
		}
	}
	return out
}
