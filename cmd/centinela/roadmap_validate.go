package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate roadmap analysis and quality artifacts",
	RunE:  runRoadmapValidate,
}

func init() {
	roadmapCmd.AddCommand(roadmapValidateCmd)
}

func runRoadmapValidate(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return fmt.Errorf("no roadmap found — define one with Claude or run centinela init")
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		return err
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Roadmap analysis and quality are valid."))
	return nil
}
