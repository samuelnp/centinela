package main

import "github.com/spf13/cobra"

// roadmapPhaseCmd is the parent for phase-level structural operations
// (add/rename/remove). It only groups the subcommands; each subcommand does the
// thin flag-parsing and delegates to internal/roadmap.
var roadmapPhaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Add, rename, or remove a roadmap phase",
}

func init() {
	roadmapCmd.AddCommand(roadmapPhaseCmd)
}
