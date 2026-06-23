package brownmap

import (
	"github.com/samuelnp/centinela/internal/reconstruct"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// GapPhaseName is the canonical name of the schedulable gap phase that holds
// net-new work: surfaces whose reconstructed behavior is unconfirmed plus
// explicit user-stated goals. Unlike Baseline it is NOT schedule-exempt — its
// features are real, plannable work.
const GapPhaseName = "Gaps"

// gapPhases builds the gap phase(s) from the TODO-bearing targets and the
// user-stated goals, in that order, preserving input order for byte-stability.
// When there is no net-new work (no TODO targets and no goals) it returns nil so
// the draft carries a Baseline phase only and the summary can hint at --goal.
func gapPhases(todos []reconstruct.Target, goals []string) []roadmap.Phase {
	features := make([]roadmap.Feature, 0, len(todos)+len(goals))
	for _, t := range todos {
		features = append(features, todoGapFeature(t))
	}
	for _, g := range goals {
		features = append(features, goalGapFeature(g))
	}
	if len(features) == 0 {
		return nil
	}
	return []roadmap.Phase{{
		Name:     GapPhaseName,
		Note:     gapNote,
		Features: features,
	}}
}

// gapNote is the gap phase blockquote distinguishing net-new work from Baseline.
const gapNote = "Net-new work: reconstructed surfaces whose behavior is " +
	"unconfirmed plus user-stated goals. Schedulable — refine and order before starting."

// todoGapFeature builds a schedulable gap feature for a TODO-bearing target.
func todoGapFeature(t reconstruct.Target) roadmap.Feature {
	return roadmap.Feature{
		Name:        t.Slug + "-confirm",
		Description: "Confirm reconstructed behavior of `" + t.Pkg + "` and replace its TODO markers.",
		Fixes:       "unconfirmed reconstructed " + string(roleOrModule(t.Role)) + " surface",
		Source:      &roadmap.Source{Feature: sourceFeature, Role: "big-thinker"},
	}
}

// goalGapFeature builds a schedulable gap feature from a user-stated goal string.
func goalGapFeature(goal string) roadmap.Feature {
	return roadmap.Feature{
		Name:        goal,
		Description: "Net-new goal supplied via --goal.",
		Fixes:       "user-stated capability gap",
		Source:      &roadmap.Source{Feature: sourceFeature, Role: "operator-goal"},
	}
}
