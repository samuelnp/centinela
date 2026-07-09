package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	phaseAddNote  string
	phaseAddAfter string
)

var roadmapPhaseAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Insert a new empty phase (before Backlog, or after --after)",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapPhaseAdd,
}

func init() {
	roadmapPhaseAddCmd.Flags().StringVar(&phaseAddNote, "note", "", "optional phase note (rationale)")
	roadmapPhaseAddCmd.Flags().StringVar(&phaseAddAfter, "after", "", "insert immediately after this phase")
	roadmapPhaseCmd.AddCommand(roadmapPhaseAddCmd)
}

func runRoadmapPhaseAdd(_ *cobra.Command, args []string) error {
	if err := roadmap.PhaseAdd(roadmap.RoadmapFile, args[0], phaseAddNote, phaseAddAfter); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Added phase %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0])))
	return nil
}
