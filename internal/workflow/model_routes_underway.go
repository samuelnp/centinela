package workflow

import (
	"os"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// RoleStepUnderway reports whether delegation for the role's scheduled step has
// begun. A pure "in-progress" status cannot answer this: the first step is
// in-progress from `start`, which is exactly the sanctioned routing window.
// Underway = the step is behind the cursor (or the workflow is done, or the
// cursor is unreadable), or it is the current step AND an evidence artifact
// from THIS run exists — `evidence init` writes both stubs, so either one
// proves delegation began.
func RoleStepUnderway(wf *Workflow, role orchestration.Role) bool {
	step, ok := RoleScheduledStep(wf, role)
	if !ok {
		return false
	}
	if wf.CurrentStep == "done" {
		return true
	}
	order := wf.OrderedSteps()
	current := stepIndexIn(wf.CurrentStep, order)
	// A currentStep outside the step order proves nothing about where the
	// workflow stands; treating it as "not underway" would disarm EVERY
	// downgrade refusal at once, so the fail-safe reading is "underway".
	if current < 0 {
		return true
	}
	if stepIndexIn(step, order) < current {
		return true
	}
	if step != wf.CurrentStep {
		return false
	}
	return roleEvidenceFromThisRun(wf, role)
}

// clockSkewGrace absorbs the gap between the precise clock StartedAt is stamped
// from and the coarser, tick-granularity clock a kernel stamps mtimes from. On
// Linux the latter can read several milliseconds EARLIER than a time.Now() taken
// just before the write, which made an artifact written moments after `start`
// look like it predated the run — green on macOS, red on CI. A stub genuinely
// left by an earlier run predates it by minutes at least, so seconds of slack
// cost this check nothing it was meant to catch.
const clockSkewGrace = 2 * time.Second

// roleEvidenceFromThisRun reports whether an evidence artifact written since the
// workflow started exists for the role. A stub whose mtime PREDATES StartedAt is
// left over from an earlier run of the same slug (a re-start, or a rewind that
// never cleaned up) and must not close the sanctioned start-time routing window.
// A workflow with no StartedAt (legacy JSON) keeps the conservative reading:
// any artifact counts.
func roleEvidenceFromThisRun(wf *Workflow, role orchestration.Role) bool {
	cutoff := wf.StartedAt.Add(-clockSkewGrace)
	for _, path := range []string{
		orchestration.MarkdownPath(wf.Feature, role),
		orchestration.JSONPath(wf.Feature, role),
	} {
		info, err := os.Stat(path)
		if err == nil && !info.ModTime().Before(cutoff) {
			return true
		}
	}
	return false
}
