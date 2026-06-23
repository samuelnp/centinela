package roadmap

import (
	"encoding/json"
	"os"

	"github.com/samuelnp/centinela/internal/workflow"
)

const RoadmapFile = ".workflow/roadmap.json"

// Feature, Phase and Roadmap struct definitions live in types.go to keep this
// file within the ≤100-line budget after the prose fields were added.

// Load reads roadmap.json from disk.
func Load() (*Roadmap, error) {
	data, err := os.ReadFile(RoadmapFile)
	if err != nil {
		return nil, err
	}
	var r Roadmap
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if err := ValidateDependencies(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Save writes roadmap.json to disk, creating .workflow/ if needed.
func Save(r *Roadmap) error {
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(RoadmapFile, data, 0644)
}

// FeatureStatus derives the current status by inspecting the workflow state.
// Returns "planned", "in-progress", or "done" — never stored in roadmap.json.
func FeatureStatus(name string) string {
	wf, err := workflow.Load(name)
	if err != nil {
		return "planned"
	}
	if wf.CurrentStep == "done" {
		return "done"
	}
	return "in-progress"
}

// Summary returns counts of features by status across all schedulable phases.
// Backlog entries (deferred findings) and Baseline entries (already-built
// capability) are not schedulable features, so they are excluded from every
// count via the shared isNonSchedulablePhase predicate.
func (r *Roadmap) Summary() (planned, inProgress, done int) {
	for _, phase := range r.Phases {
		if isNonSchedulablePhase(phase.Name) {
			continue
		}
		for _, f := range phase.Features {
			switch FeatureStatus(f.Name) {
			case "done":
				done++
			case "in-progress":
				inProgress++
			default:
				planned++
			}
		}
	}
	return
}
