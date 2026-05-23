package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmapcheckpoint"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapIterateCmd = &cobra.Command{
	Use:   "iterate",
	Short: "Record the choice to keep iterating on the roadmap definition",
	RunE:  runRoadmapIterate,
}

func init() {
	roadmapCmd.AddCommand(roadmapIterateCmd)
}

func runRoadmapIterate(_ *cobra.Command, _ []string) error {
	if err := roadmapcheckpoint.WriteMarker(roadmapcheckpoint.MarkerPath, time.Now()); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Roadmap iteration marker written; checkpoint prompt suppressed until artifacts change."))
	return nil
}
