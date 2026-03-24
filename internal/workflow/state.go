package workflow

import (
	"encoding/json"
	"fmt"
	"os"
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
	Feature     string               `json:"feature"`
	StartedAt   time.Time            `json:"startedAt"`
	CurrentStep string               `json:"currentStep"`
	Steps       map[string]StepState `json:"steps"`
	StepOrder   []string             `json:"stepOrder,omitempty"`
}

// WorkflowDir is the directory where workflow JSON files are stored.
const WorkflowDir = ".workflow"

// FilePath returns the JSON file path for a given feature.
func FilePath(feature string) string {
	return filepath.Join(WorkflowDir, feature+".json")
}

// Load reads and parses a workflow file from disk.
func Load(feature string) (*Workflow, error) {
	data, err := os.ReadFile(FilePath(feature))
	if err != nil {
		return nil, fmt.Errorf("no workflow found for %q", feature)
	}
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("invalid workflow file: %w", err)
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

// New creates a fresh workflow starting at the "plan" step.
func New(feature string) *Workflow {
	return NewWithOrder(feature, DefaultStepOrder)
}
