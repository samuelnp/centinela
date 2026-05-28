package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceRepairCmd = &cobra.Command{
	Use:   "repair <feature>",
	Short: "Remove orphaned .json.tmp files from crashed writes (idempotent)",
	Args:  cobra.ExactArgs(1),
	RunE:  runEvidenceRepair,
}

func init() {
	evidenceCmd.AddCommand(evidenceRepairCmd)
}

func runEvidenceRepair(_ *cobra.Command, args []string) error {
	feature := args[0]
	removed, err := evidence.Repair(feature)
	if err != nil {
		return err
	}
	if len(removed) == 0 {
		fmt.Fprintln(os.Stdout, "no orphaned temp files found")
		return nil
	}
	for _, p := range removed {
		fmt.Fprintf(os.Stdout, "removed %s\n", p)
	}
	return nil
}
