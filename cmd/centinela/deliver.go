package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/gitutil"
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

var deliverVia string

var deliverCmd = &cobra.Command{
	Use:   "deliver <feature>",
	Short: "Deliver a completed feature via the explicitly chosen path",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeliver,
}

func init() {
	deliverCmd.Flags().StringVar(&deliverVia, "via", "", "Delivery path: pr|merge (required)")
	_ = deliverCmd.MarkFlagRequired("via")
	rootCmd.AddCommand(deliverCmd)
}

// runDeliver is a thin orchestrator: gitutil decides which paths the live
// repo supports and runDeliver dispatches the user's explicit --via choice.
// No matrix logic lives here (G7).
func runDeliver(cmd *cobra.Command, args []string) error {
	feature := args[0]
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}
	via := gitutil.Option(deliverVia)
	if via != gitutil.OptionPR && via != gitutil.OptionMerge {
		return fmt.Errorf("choose --via pr|merge")
	}

	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}
	hasOrigin, _ := gitutil.HasOriginRemote(".")
	opts := gitutil.DeliveryOptions(hasOrigin, wf.WorktreePath != "")
	if !gitutil.Supports(opts, via) {
		if via == gitutil.OptionPR {
			return fmt.Errorf("no origin remote — PR delivery unavailable for %q", feature)
		}
		return fmt.Errorf("worktree mode required — merge delivery unavailable for %q", feature)
	}

	if via == gitutil.OptionMerge {
		return runMerge(cmd, []string{feature})
	}
	return runDeliverPR(cmd, feature)
}
