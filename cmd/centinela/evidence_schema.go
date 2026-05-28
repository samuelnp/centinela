package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceSchemaCmd = &cobra.Command{
	Use:   "schema <role>",
	Short: "Print the JSON skeleton for a role (for prompt embedding)",
	Args:  cobra.ExactArgs(1),
	RunE:  runEvidenceSchema,
}

func init() {
	evidenceCmd.AddCommand(evidenceSchemaCmd)
}

func runEvidenceSchema(_ *cobra.Command, args []string) error {
	role, err := evidence.ParseRole(args[0])
	if err != nil {
		return err
	}
	out, err := evidence.SchemaSkeleton(role, Version)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
