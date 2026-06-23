package brownmap

import (
	"github.com/samuelnp/centinela/internal/reconstruct"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// sourceFeature names the originating workflow on every generated feature's
// Source provenance, so a reviewer can trace the draft back to its generator.
const sourceFeature = "brownfield-roadmap-generation"

// baselinePhase builds the schedule-exempt Baseline phase from the
// reconstructor's already-sorted targets: one roadmap.Feature per target,
// preserving target order so the phase is byte-stable. An empty target list
// yields an empty (non-nil) Baseline phase so the draft is never malformed.
func baselinePhase(targets []reconstruct.Target) roadmap.Phase {
	p := roadmap.Phase{
		Name:     roadmap.BaselinePhaseName,
		Note:     baselineNote,
		Features: make([]roadmap.Feature, 0, len(targets)),
	}
	for _, t := range targets {
		p.Features = append(p.Features, roadmap.Feature{
			Name:        t.Slug,
			Description: baselineDescription(t),
			Source:      &roadmap.Source{Feature: sourceFeature, Role: "big-thinker"},
		})
	}
	return p
}

// baselineNote is the phase blockquote explaining the Baseline phase is
// already-built capability, not schedulable work.
const baselineNote = "Already-built surfaces detected from the codebase inventory. " +
	"These document existing capability and are never re-planned."

// baselineDescription renders the one-line human-facing prose for a baseline
// feature from its reconstructed role and selection reason.
func baselineDescription(t reconstruct.Target) string {
	return "Already-built " + string(roleOrModule(t.Role)) + " surface `" + t.Pkg + "` — " + t.Reason + "."
}

// roleOrModule normalizes an empty role to the module label for display, so a
// target with an unknown role still renders a non-empty description.
func roleOrModule(r reconstruct.Role) reconstruct.Role {
	if r == "" {
		return reconstruct.RoleModule
	}
	return r
}
