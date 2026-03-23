package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var completeCmd = &cobra.Command{
	Use:   "complete <feature>",
	Short: "Complete the current step and advance to the next",
	Args:  cobra.ExactArgs(1),
	RunE:  runComplete,
}

var saveWorkflow = workflow.Save

func init() {
	rootCmd.AddCommand(completeCmd)
}

func runComplete(_ *cobra.Command, args []string) error {
	feature := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}

	if wf.CurrentStep == "done" {
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Workflow for %q is already complete.", feature)))
		return nil
	}

	current := wf.CurrentStep

	// Validate step requires all gates to pass before advancing.
	if current == "validate" {
		if err := executeValidation(); err != nil {
			return err
		}
	}

	if err := wf.Complete(cfg); err != nil {
		return err
	}
	if err := saveWorkflow(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	if !cfg.Workflow.DisableAutoCommit {
		commitStep(feature, current, workflow.StepNumber(current))
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Step %q completed for %q.", current, feature)))
	if wf.CurrentStep == "done" {
		fmt.Println(ui.StyleGreen.Bold(true).Render(fmt.Sprintf("Workflow complete for %q!", feature)))
	} else {
		fmt.Println(ui.RenderStep("Next step", wf.CurrentStep))
	}
	if warn := workflow.ProductionReadinessWarning(feature, cfg); warn != "" {
		fmt.Println(ui.RenderProductionReadinessWarning(feature))
	}
	return nil
}
