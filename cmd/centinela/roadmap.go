package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapCmd = &cobra.Command{
	Use:   "roadmap",
	Short: "Show roadmap status and feature completion across all phases",
	RunE:  runRoadmap,
}

func init() {
	rootCmd.AddCommand(roadmapCmd)
}

func runRoadmap(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	fmt.Println(ui.RenderRoadmap(r))
	return nil
}
