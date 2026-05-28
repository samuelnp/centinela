package main

import (
	"github.com/spf13/cobra"
)

// evidenceCmd is the thin orchestrator for `centinela evidence`. All
// domain logic — schema, IO, validation, fix hints — lives in
// internal/evidence per G7. This file ONLY wires subcommands.
var evidenceCmd = &cobra.Command{
	Use:   "evidence",
	Short: "Author, inspect, and validate .workflow/<feature>-<role>.json evidence",
	Long: "Typed CLI for the orchestration evidence contract. Replaces\n" +
		"hand-written JSON via python/jq/heredoc. See\n" +
		"docs/architecture/evidence-contract.md.",
}

func init() {
	rootCmd.AddCommand(evidenceCmd)
}
