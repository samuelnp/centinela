package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var completeCmd = &cobra.Command{
	Use:   "complete <feature>",
	Short: "Complete the current step and advance to the next",
	Args:  cobra.ExactArgs(1),
	RunE:  runComplete,
}

func init() {
	rootCmd.AddCommand(completeCmd)
}

func runComplete(_ *cobra.Command, args []string) error {
	feature := args[0]

	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}

	if wf.CurrentStep == "done" {
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Workflow for %q is already complete.", feature)))
		return nil
	}

	current := wf.CurrentStep

	if err := wf.Complete(); err != nil {
		return err
	}
	if err := workflow.Save(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Step %q completed for %q.", current, feature)))
	if wf.CurrentStep == "done" {
		fmt.Println(ui.StyleGreen.Bold(true).Render(fmt.Sprintf("Workflow complete for %q!", feature)))
	} else {
		fmt.Println(ui.RenderStep("Next step", wf.CurrentStep))
	}
	return nil
}
