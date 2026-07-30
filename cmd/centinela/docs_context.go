package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/docsctx"
	"github.com/samuelnp/centinela/internal/worktree"
)

var docsContextCmd = &cobra.Command{
	Use:           "context <feature>",
	Short:         "Print the curated docs-step inputs for a feature as plain markdown",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runDocsContext,
}

func init() {
	docsCmd.AddCommand(docsContextCmd)
}

func runDocsContext(_ *cobra.Command, args []string) error {
	feature := args[0]
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}
	ctx, err := docsctx.Load(feature)
	if err != nil {
		return err
	}
	fmt.Print(docsctx.Render(ctx))
	return nil
}
