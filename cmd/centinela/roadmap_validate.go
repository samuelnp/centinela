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
		return roadmapCommandError(err)
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		return err
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Roadmap analysis and quality are valid."))
	printQualityAdvisories()
	return nil
}

// printQualityAdvisories lists low self-assigned quality scores as advice on an
// otherwise valid roadmap. It cannot change the exit code: the overall >= 9
// refusal was deleted outright, in every profile, and nothing replaced it.
func printQualityAdvisories() {
	advisories := roadmap.QualityAdvisories()
	if len(advisories) == 0 {
		return
	}
	fmt.Println(ui.StyleMuted.Render("Advisory — low self-assigned quality scores:"))
	for _, line := range advisories {
		fmt.Println(ui.StyleMuted.Render("  · " + line))
	}
}
