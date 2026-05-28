package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceValidateCmd = &cobra.Command{
	Use:   "validate <feature>",
	Short: "Validate every evidence file for a feature; non-zero exit on issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runEvidenceValidate,
}

func init() {
	evidenceCmd.AddCommand(evidenceValidateCmd)
}

func runEvidenceValidate(_ *cobra.Command, args []string) error {
	feature := args[0]
	hints := evidence.ValidateFeature(feature, nil)
	if len(hints) == 0 {
		fmt.Fprintf(os.Stdout, "evidence ok for %q\n", feature)
		return nil
	}
	for _, h := range hints {
		fmt.Fprintln(os.Stderr, h.String())
	}
	return fmt.Errorf("evidence validation failed: %d issue(s)", len(hints))
}
