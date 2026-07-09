package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	reorderBefore string
	reorderAfter  string
)

var roadmapReorderCmd = &cobra.Command{
	Use:   "reorder <slug>",
	Short: "Reposition a feature relative to a sibling",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapReorder,
}

func init() {
	roadmapReorderCmd.Flags().StringVar(&reorderBefore, "before", "", "reposition before this sibling feature")
	roadmapReorderCmd.Flags().StringVar(&reorderAfter, "after", "", "reposition after this sibling feature")
	roadmapCmd.AddCommand(roadmapReorderCmd)
}

func runRoadmapReorder(_ *cobra.Command, args []string) error {
	if reorderBefore != "" && reorderAfter != "" {
		return fmt.Errorf("--before and --after are mutually exclusive")
	}
	if reorderBefore == "" && reorderAfter == "" {
		return fmt.Errorf("--before or --after is required")
	}
	if err := roadmap.Reorder(roadmap.RoadmapFile, roadmap.ReorderRequest{
		Slug:         args[0],
		BeforeAnchor: reorderBefore,
		AfterAnchor:  reorderAfter,
	}); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Reordered %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0])))
	return nil
}
