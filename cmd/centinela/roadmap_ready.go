package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List features that are ready to start right now (all dependencies done)",
	RunE:  runRoadmapReady,
}

func init() {
	roadmapCmd.AddCommand(roadmapReadyCmd)
}

func runRoadmapReady(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	ready := roadmap.ReadySet(r)
	fmt.Println(ui.RenderReadyList(ready))
	return nil
}
