// Package teamdashboard computes a read-only, multi-feature team-status board
// from in-memory inputs the caller reads off disk. It is an aggregator over the
// workflow + roadmap domains and the telemetry + insights aggregator: it imports
// only those (read-only) plus stdlib, never cmd/ or internal/ui, and performs no
// I/O and no git. Compute is pure and deterministic — every emitted list follows
// a stable order (active-workflow order, roadmap file order, insights ranking);
// no map is ranged in output order.
package teamdashboard

import (
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Inputs is the plain, caller-populated aggregate. The package reads nothing
// from disk: cmd/ fills every field. Now is injected for deterministic ages.
type Inputs struct {
	Active  []*workflow.Workflow // workflow.ActiveWorkflows(workflow.WorkflowDir)
	Roadmap *roadmap.Roadmap     // roadmap.Load() result, nil when absent/unreadable
	Events  []telemetry.Event    // telemetry.ReadDefault() result
	Owners  map[string]string    // feature -> git-derived owner ("unknown" allowed)
	Now     time.Time            // age reference; cmd/ passes time.Now().UTC()
}

// Dashboard is the pure, serializable board. Field names are a stable --json
// contract; do not rename without bumping consumers.
type Dashboard struct {
	Features []FeatureRow    // one row per active workflow (input order preserved)
	Roadmap  RoadmapBurndown // schedulable-phase counts + overall
	Gates    []GateHealth    // gate-failure tallies, ranked desc/asc
}

// FeatureRow is one in-flight feature. Blank Profile/Archetype/Worktree are
// surfaced verbatim here; the renderer fills "default"/"canonical"/"—".
type FeatureRow struct {
	Feature   string // wf.Feature
	Step      string // wf.CurrentStep
	StepIndex int    // done-count (0-based position in OrderedSteps)
	StepTotal int    // len(wf.OrderedSteps())
	AgeDays   int    // floor((Now - wf.StartedAt)/24h); 0 if StartedAt zero/future
	Profile   string // wf.EnforcementProfile
	Archetype string // wf.Archetype
	Worktree  string // wf.WorktreePath
	Owner     string // Inputs.Owners[feature]; "unknown" when missing
}

// RoadmapBurndown is the schedulable-only roadmap progress. Present is false
// when Inputs.Roadmap == nil (the empty state).
type RoadmapBurndown struct {
	Present    bool          // false when Inputs.Roadmap == nil
	Planned    int           // from Roadmap.Summary()
	InProgress int           // from Roadmap.Summary()
	Done       int           // from Roadmap.Summary()
	Total      int           // Planned + InProgress + Done (schedulable only)
	Phases     []PhaseStatus // per-schedulable-phase done/total, file order
}

// PhaseStatus is one schedulable phase's done/total feature count.
type PhaseStatus struct {
	Name  string
	Done  int
	Total int
}

// GateHealth is one gate's failure tally ("<none>" bucket inherited from
// insights.Gates).
type GateHealth struct {
	Gate  string
	Fails int
}
