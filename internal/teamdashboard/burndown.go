package teamdashboard

import "github.com/samuelnp/centinela/internal/roadmap"

// burndown derives the schedulable-only roadmap progress. A nil roadmap (absent
// or unreadable on disk) yields the empty state {Present:false}. Otherwise the
// overall counts come from Roadmap.Summary() (which already excludes the
// non-schedulable Backlog/Baseline phases) and each schedulable phase gets a
// PhaseStatus in roadmap file order, with Done = features whose FeatureStatus
// is "done".
func burndown(r *roadmap.Roadmap) RoadmapBurndown {
	if r == nil {
		return RoadmapBurndown{Present: false}
	}
	planned, inProgress, done := r.Summary()
	b := RoadmapBurndown{
		Present:    true,
		Planned:    planned,
		InProgress: inProgress,
		Done:       done,
		Total:      planned + inProgress + done,
	}
	for _, phase := range r.Phases {
		if roadmap.IsBacklogPhaseName(phase.Name) || roadmap.IsBaselinePhaseName(phase.Name) {
			continue
		}
		ps := PhaseStatus{Name: phase.Name, Total: len(phase.Features)}
		for _, f := range phase.Features {
			if roadmap.FeatureStatus(f.Name) == "done" {
				ps.Done++
			}
		}
		b.Phases = append(b.Phases, ps)
	}
	return b
}
