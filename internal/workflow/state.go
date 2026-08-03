package workflow

import (
	"path/filepath"
	"time"
)

// StepState holds the status and completion time of a single workflow step.
type StepState struct {
	Status      string  `json:"status"`
	CompletedAt *string `json:"completedAt"`
}

// Workflow represents the full state of a feature workflow.
type Workflow struct {
	// SchemaVersion is the state-file schema this file was written against. It
	// is first so it is the first key in the marshalled JSON — `head -3` of any
	// state file shows it — and carries no omitempty so every file this binary
	// writes is stamped. Absent on disk means version 1. A HIGHER value is
	// refused by Save and never by Load; see SchemaVersion for why.
	SchemaVersion     int                  `json:"schemaVersion"`
	Feature           string               `json:"feature"`
	StartedAt         time.Time            `json:"startedAt"`
	CurrentStep       string               `json:"currentStep"`
	Steps             map[string]StepState `json:"steps"`
	StepOrder         []string             `json:"stepOrder,omitempty"`
	OrchestrationMode string               `json:"orchestrationMode,omitempty"`
	// EnforcementProfile pins the feature's profile at start. Empty on workflows
	// created before this field existed — EffectiveProfile falls back to the
	// global config / strict default for those.
	EnforcementProfile string `json:"enforcementProfile,omitempty"`
	// Archetype pins the feature's workflow track at start (canonical, hotfix,
	// refactor, spike). Empty on workflows created before this field existed —
	// DisplayArchetype falls back to canonical for those.
	Archetype string `json:"archetype,omitempty"`
	// WorktreePath is the absolute path of the worktree this workflow runs
	// in. Empty when the project uses a single-checkout flow.
	WorktreePath string `json:"worktreePath,omitempty"`
	// DriverModel is the model id pinned at start that keys this feature's default
	// enforcement profile via the capability tier. Empty on workflows created
	// before this field existed and when no driver model was configured —
	// EffectiveProfile simply skips the capability tier for those.
	DriverModel string `json:"driverModel,omitempty"`
	// ValidateContract pins the validate-step gate contract at start. Empty on
	// workflows created before this field existed — those keep the historical
	// existence-only gatekeeper check (see UsesAdversarialVerifier).
	ValidateContract string `json:"validateContract,omitempty"`
	// PlanContract pins the plan-step role contract at start. Empty on workflows
	// created before this field existed — those keep demanding the complete
	// legacy big-thinker + feature-specialist pair (see UsesUnifiedPlanner).
	PlanContract string `json:"planContract,omitempty"`
	// Revisions is the append-only audit log of backward transitions performed by
	// `centinela revise`. Empty/absent on workflows that were never rewound
	// (back-compat, like Archetype) — RevisionsSummary handles the empty case.
	// The Revision type lives in rewind.go alongside the logic that appends it.
	Revisions []Revision `json:"revisions,omitempty"`
	// ModelRoutes is the per-feature role→routing decision recorded by
	// `centinela route set` under dynamic routing. Absent/omitted on every
	// workflow written before this field existed and on every static-mode
	// workflow — a nil map simply means "no routes" (see model_routes.go).
	ModelRoutes map[string]ModelRoute `json:"modelRoutes,omitempty"`

	// loadedDigest fingerprints the raw bytes Load read, so Save can refuse to
	// overwrite a file another process has changed since. It is UNEXPORTED on
	// purpose: encoding/json ignores it, so the on-disk shape is untouched and
	// not one existing fixture needs to change. Empty on a workflow built by
	// New/NewWithOrder, which is what exempts `start` and the autostart hook
	// from a conflict check they cannot meaningfully fail.
	loadedDigest string
}

// WorkflowDir is the directory where workflow JSON files are stored.
const WorkflowDir = ".workflow"

// FilePath returns the JSON file path for a given feature.
func FilePath(feature string) string {
	return filepath.Join(WorkflowDir, feature+".json")
}

// Load and Save live in state_io.go: the durability rules (atomic replace,
// schema version, stale-write detection) are a concern of their own and this
// file is at the 100-line cap.
