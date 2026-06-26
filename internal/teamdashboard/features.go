package teamdashboard

import (
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

// features builds one FeatureRow per active workflow, preserving the input order
// (ActiveWorkflows is already deduped and sorted mtime-descending). It never
// touches disk or git: owners come from the caller-supplied map.
func features(active []*workflow.Workflow, owners map[string]string, now time.Time) []FeatureRow {
	rows := make([]FeatureRow, 0, len(active))
	for _, wf := range active {
		if wf == nil {
			continue
		}
		rows = append(rows, FeatureRow{
			Feature:   wf.Feature,
			Step:      wf.CurrentStep,
			StepIndex: doneCount(wf),
			StepTotal: len(wf.OrderedSteps()),
			AgeDays:   ageDays(wf.StartedAt, now),
			Profile:   wf.EnforcementProfile,
			Archetype: wf.Archetype,
			Worktree:  wf.WorktreePath,
			Owner:     ownerOf(owners, wf.Feature),
		})
	}
	return rows
}

// doneCount is the 0-based position of CurrentStep in the ordered step list —
// the count of steps already completed. Mirrors ui.wfDoneCount: "done" means all
// steps complete; an unknown CurrentStep counts as zero.
func doneCount(wf *workflow.Workflow) int {
	steps := wf.OrderedSteps()
	if wf.CurrentStep == "done" {
		return len(steps)
	}
	for i, step := range steps {
		if step == wf.CurrentStep {
			return i
		}
	}
	return 0
}

// ageDays is floor((now - started)/24h). A zero or future StartedAt yields 0 so
// freshly-created or clock-skewed workflows never show a negative age.
func ageDays(started, now time.Time) int {
	if started.IsZero() || !started.Before(now) {
		return 0
	}
	return int(now.Sub(started) / (24 * time.Hour))
}

// ownerOf returns owners[feature], defaulting a missing or empty entry to
// "unknown" so the column is always populated (owner is advisory).
func ownerOf(owners map[string]string, feature string) string {
	if o, ok := owners[feature]; ok && o != "" {
		return o
	}
	return "unknown"
}
