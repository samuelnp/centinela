package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:           "docs",
	Short:         "Docs-step helpers for markdown-first documentation",
	SilenceUsage:  true,
	SilenceErrors: true,
	// Deleted subcommands (generate, validate) must fail loudly as unknown
	// commands instead of printing help with exit 0.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
