package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var addPhase string
var addDescription string
var addArchetype string
var addDependsOn []string

var roadmapAddCmd = &cobra.Command{
	Use:   "add <slug>",
	Short: "Author a new draft feature directly into a schedulable phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapAdd,
}

func init() {
	roadmapAddCmd.Flags().StringVar(&addPhase, "phase", "", "target schedulable (non-Backlog) phase (required)")
	roadmapAddCmd.Flags().StringVar(&addDescription, "description", "", "human-facing bullet prose")
	roadmapAddCmd.Flags().StringVar(&addArchetype, "archetype", "", "workflow archetype (canonical, hotfix, refactor, spike)")
	roadmapAddCmd.Flags().StringSliceVar(&addDependsOn, "depends-on", nil, "feature slugs this feature depends on")
	roadmapCmd.AddCommand(roadmapAddCmd)
}

func runRoadmapAdd(_ *cobra.Command, args []string) error {
	if addPhase == "" {
		return fmt.Errorf("--phase is required")
	}
	if err := roadmap.Add(roadmap.RoadmapFile, roadmap.AddRequest{
		Slug:        args[0],
		Phase:       addPhase,
		Description: addDescription,
		Archetype:   addArchetype,
		DependsOn:   addDependsOn,
	}); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Added draft %q to %q. Score it with roadmap promote to finalize.", args[0], addPhase)))
	return nil
}
