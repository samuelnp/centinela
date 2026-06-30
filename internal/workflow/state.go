package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// StepState holds the status and completion time of a single workflow step.
type StepState struct {
	Status      string  `json:"status"`
	CompletedAt *string `json:"completedAt"`
}

// Workflow represents the full state of a feature workflow.
type Workflow struct {
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
	// Revisions is the append-only audit log of backward transitions performed by
	// `centinela revise`. Empty/absent on workflows that were never rewound
	// (back-compat, like Archetype) — RevisionsSummary handles the empty case.
	// The Revision type lives in rewind.go alongside the logic that appends it.
	Revisions []Revision `json:"revisions,omitempty"`
}

// WorkflowDir is the directory where workflow JSON files are stored.
const WorkflowDir = ".workflow"

// FilePath returns the JSON file path for a given feature.
func FilePath(feature string) string {
	return filepath.Join(WorkflowDir, feature+".json")
}

// Load reads and parses a workflow file from disk. Only a genuinely
// missing state file reports absence; read and parse failures surface
// with the state file path so they are never mistaken for "not started".
func Load(feature string) (*Workflow, error) {
	path := FilePath(feature)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no workflow found for %q", feature)
		}
		return nil, fmt.Errorf("reading workflow file %s: %w", path, err)
	}
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("invalid workflow file %s: %w", path, err)
	}
	return &wf, nil
}

// Save writes a workflow to disk.
func Save(wf *Workflow) error {
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(FilePath(wf.Feature), data, 0644)
}

// New creates a fresh workflow starting at the "plan" step under the strict
// (back-compat default) profile.
func New(feature string) *Workflow {
	return NewWithOrder(feature, DefaultStepOrder, config.ProfileStrict)
}
