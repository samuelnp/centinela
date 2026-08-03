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

  Feature resolved (CWD is inside a .worktrees/<feature> checkout whose workflow
  state exists, or exactly one workflow is active): "feature" and "handoffTo"
  are filled in exactly as "centinela evidence init <feature> <role>" would, so
  the printed handoffTo is the value the completion gate accepts. A worktree
  checkout resolves from ANY depth inside it — its state is read from the
  worktree root, not from the directory you happen to stand in.

  No feature resolved (no worktree; or a worktree with no workflow state; or
  zero, or two or more, active workflows): it never guesses. "feature" prints
  <feature-slug> and "handoffTo" prints <successor-role> — both slots for you to
  fill. Pasted verbatim into real evidence for a role that workflow's contract
  REQUIRES, that handoffTo is refused by the chain gate, which names the true
  successor and the exact "centinela evidence set" command to fix it. For a role
  the contract does not require, the chain gate does not inspect handoffTo at
  all, so the placeholder passes "evidence validate" — it is a slot marked for a
  human to fill, not a value the gate promises to catch everywhere.`,
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
	feature, root := evidence.ResolveActiveFeature(cwd)
	var out []byte
	// Derivation reads every input through a CWD-relative path, so it must run
	// where the state lives: the worktree root when the CWD resolved one (root
	// == "" means the CWD already is that place).
	render := func() error {
		var rerr error
		out, rerr = evidence.SchemaSkeleton(feature, role, Version)
		return rerr
	}
	if root != "" {
		err = inDir(root, render)
	} else {
		err = render()
	}
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
