package workflow

import (
	"fmt"
	"strings"
	"time"
)

// Revision records a single backward transition: the step the workflow was on
// (From), the step it was rewound to (To), the human reason, and when. It is
// append-only — RewindTo never mutates or removes an existing entry.
type Revision struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Reason string    `json:"reason"`
	At     time.Time `json:"at"`
}

// RewindTo performs a controlled backward transition: it re-opens every step
// strictly after target to "pending", sets target itself to "in-progress",
// clears the CompletedAt of every touched step, moves CurrentStep to target,
// and appends a Revision to the audit log. It returns the names of the steps
// that were re-opened (every step after target) so the caller knows whose
// certification evidence to invalidate. It is pure state — no I/O, no evidence.
//
// The transition is rejected (state untouched) unless ALL hold, each error
// naming the offending value: target is a real step for THIS feature's order;
// target is strictly before CurrentStep (revise is backward-only); the
// workflow is not already "done"; reason is non-empty after trimming.
func (wf *Workflow) RewindTo(target, reason string) ([]string, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("revise reason must not be empty")
	}
	if wf.CurrentStep == "done" {
		return nil, fmt.Errorf("cannot revise a completed workflow")
	}
	order := wf.OrderedSteps()
	targetIdx := stepIndexIn(target, order)
	if targetIdx < 0 {
		return nil, fmt.Errorf("unrecognised step %q (valid: %v)", target, order)
	}
	currentIdx := stepIndexIn(wf.CurrentStep, order)
	if targetIdx >= currentIdx {
		return nil, fmt.Errorf("target step %q is not strictly before the current step %q", target, wf.CurrentStep)
	}

	from := wf.CurrentStep
	reopened := reopenedSteps(order, target)
	for _, step := range reopened {
		s := wf.Steps[step]
		s.Status = "pending"
		s.CompletedAt = nil
		wf.Steps[step] = s
	}
	t := wf.Steps[target]
	t.Status = "in-progress"
	t.CompletedAt = nil
	wf.Steps[target] = t

	wf.CurrentStep = target
	wf.Revisions = append(wf.Revisions, Revision{
		From:   from,
		To:     target,
		Reason: strings.TrimSpace(reason),
		At:     time.Now().UTC(),
	})
	return reopened, nil
}

// RevisionsSummary renders the audit count plus the latest reason for the
// status view, or "" when the workflow was never rewound. The display logic
// lives here so internal/ui stays logic-free (mirrors DisplayArchetype).
func RevisionsSummary(wf *Workflow) string {
	if wf == nil || len(wf.Revisions) == 0 {
		return ""
	}
	last := wf.Revisions[len(wf.Revisions)-1]
	return fmt.Sprintf("%d  (last: %q)", len(wf.Revisions), last.Reason)
}

// reopenedSteps returns every step strictly after target in order. It is
// archetype-aware because it operates on the workflow's own step order, never
// DefaultStepOrder. Target absent yields an empty slice.
func reopenedSteps(order []string, target string) []string {
	idx := stepIndexIn(target, order)
	if idx < 0 {
		return nil
	}
	out := make([]string, 0, len(order)-idx-1)
	out = append(out, order[idx+1:]...)
	return out
}
