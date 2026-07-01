package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapRemoveCmd = &cobra.Command{
	Use:     "remove <slug>",
	Aliases: []string{"rm"},
	Short:   "Delete a planned feature (refused if depended-on or in-progress/done)",
	Args:    cobra.ExactArgs(1),
	RunE:    runRoadmapRemove,
}

func init() {
	roadmapCmd.AddCommand(roadmapRemoveCmd)
}

func runRoadmapRemove(_ *cobra.Command, args []string) error {
	if err := roadmap.Remove(roadmap.RoadmapFile, args[0]); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Removed %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0])))
	return nil
}
