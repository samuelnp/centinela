package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var phaseRemoveForce bool

var roadmapPhaseRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Delete an empty phase (non-empty refused unless --force)",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapPhaseRemove,
}

func init() {
	roadmapPhaseRemoveCmd.Flags().BoolVar(&phaseRemoveForce, "force", false,
		"also remove the phase's features and their analysis/quality entries")
	roadmapPhaseCmd.AddCommand(roadmapPhaseRemoveCmd)
}

func runRoadmapPhaseRemove(_ *cobra.Command, args []string) error {
	if err := roadmap.PhaseRemove(roadmap.RoadmapFile, args[0], phaseRemoveForce); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Removed phase %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0])))
	return nil
}
