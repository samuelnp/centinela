package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	moveToPhase string
	moveBefore  string
	moveAfter   string
)

var roadmapMoveCmd = &cobra.Command{
	Use:   "move <slug>",
	Short: "Relocate a feature to another schedulable phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapMove,
}

func init() {
	roadmapMoveCmd.Flags().StringVar(&moveToPhase, "to-phase", "", "target schedulable phase (required)")
	roadmapMoveCmd.Flags().StringVar(&moveBefore, "before", "", "insert before this sibling feature")
	roadmapMoveCmd.Flags().StringVar(&moveAfter, "after", "", "insert after this sibling feature")
	roadmapCmd.AddCommand(roadmapMoveCmd)
}

func runRoadmapMove(_ *cobra.Command, args []string) error {
	if moveToPhase == "" {
		return fmt.Errorf("--to-phase is required")
	}
	if moveBefore != "" && moveAfter != "" {
		return fmt.Errorf("--before and --after are mutually exclusive")
	}
	if err := roadmap.Move(roadmap.RoadmapFile, roadmap.MoveRequest{
		Slug:         args[0],
		ToPhase:      moveToPhase,
		BeforeAnchor: moveBefore,
		AfterAnchor:  moveAfter,
	}); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Moved %q to %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0], moveToPhase)))
	return nil
}
