package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapPhaseRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a phase in place, leaving its features untouched",
	Args:  cobra.ExactArgs(2),
	RunE:  runRoadmapPhaseRename,
}

func init() {
	roadmapPhaseCmd.AddCommand(roadmapPhaseRenameCmd)
}

func runRoadmapPhaseRename(_ *cobra.Command, args []string) error {
	if err := roadmap.PhaseRename(roadmap.RoadmapFile, args[0], args[1]); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Renamed phase %q to %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0], args[1])))
	return nil
}
