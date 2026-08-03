package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceSchemaCmd = &cobra.Command{
	Use:   "schema <role>",
	Short: "Print the JSON skeleton for a role, for the feature the CWD resolves",
	Long: `Print the JSON evidence skeleton for a role, for prompt embedding.

The feature is derived from the current directory — no argument for it. Which
of the two branches applies:

  Feature resolved (CWD is inside a .worktrees/<feature> checkout, or exactly
  one workflow is active): "feature" and "handoffTo" are filled in exactly as
  "centinela evidence init <feature> <role>" would, so the printed handoffTo is
  the value the completion gate accepts.

  No feature resolved (no worktree and zero, or two or more, active workflows):
  it never guesses. "feature" prints <feature-slug> and "handoffTo" prints
  <successor-role> — both slots for you to fill. Pasted verbatim into real
  evidence, that handoffTo is refused by the chain gate, which names the true
  successor and the exact "centinela evidence set" command to fix it.`,
	Args: cobra.ExactArgs(1),
	RunE: runEvidenceSchema,
}

func init() {
	evidenceCmd.AddCommand(evidenceSchemaCmd)
}

func runEvidenceSchema(_ *cobra.Command, args []string) error {
	role, err := evidence.ParseRole(args[0])
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	out, err := evidence.SchemaSkeleton(evidence.ResolveActiveFeature(cwd), role, Version)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
