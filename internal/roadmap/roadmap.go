package roadmap

import (
	"encoding/json"
	"os"

	"github.com/samuelnp/centinela/internal/workflow"
)

const RoadmapFile = ".workflow/roadmap.json"

// Feature is a single deliverable within a phase.
type Feature struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"dependsOn,omitempty"`
	// Archetype optionally pins the workflow track for this feature; an explicit
	// --archetype flag at start overrides it. Empty resolves to canonical.
	Archetype string `json:"archetype,omitempty"`
}

// Phase groups related features under a milestone.
type Phase struct {
	Name     string    `json:"name"`
	Features []Feature `json:"features"`
}

// Roadmap holds the full project plan.
type Roadmap struct {
	Phases []Phase `json:"phases"`
}

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

// Summary returns counts of features by status across all phases.
func (r *Roadmap) Summary() (planned, inProgress, done int) {
	for _, phase := range r.Phases {
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
