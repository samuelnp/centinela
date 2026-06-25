package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitutil"
	"github.com/samuelnp/centinela/internal/memory"
	"github.com/samuelnp/centinela/internal/telemetry"
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
	model := resolveEmitModel(wf, cfg)

	// Validate step requires all gates to pass before advancing. Verification is
	// CONSTANT across every profile — NO profile branch belongs here; profiles
	// scale process, never proof.
	if current == "validate" {
		if err := executeValidation(); err != nil {
			telemetry.RecordCompleteRejected(cfg, feature, current, "gates", model)
			return err
		}
		if err := runClaimVerification(feature, current, model, cfg); err != nil {
			telemetry.RecordCompleteRejected(cfg, feature, current, "verify", model)
			return err
		}
	}

	if err := wf.Complete(cfg); err != nil {
		return err
	}
	if err := saveWorkflow(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	// Harvest the just-completed step's artifact into the memory ledger.
	// Capture is non-blocking: failures warn but never fail the advance.
	memory.Capture(feature, current, cfg)
	telemetry.RecordStepAdvanced(cfg, feature, current, model)

	if !cfg.Workflow.DisableAutoCommit {
		commitStep(feature, current, workflow.StepNumberFor(wf, current), len(wf.OrderedSteps()))
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Step %q completed for %q.", current, feature)))
	if wf.CurrentStep == "done" {
		fmt.Println(ui.StyleGreen.Bold(true).Render(fmt.Sprintf("Workflow complete for %q!", feature)))
		// Surface delivery options as guidance only — never push or merge.
		// A HasOriginRemote error is treated as "no origin"; it must never
		// block an otherwise-complete workflow.
		hasOrigin, _ := gitutil.HasOriginRemote(".")
		worktreeMode := wf.WorktreePath != ""
		opts := gitutil.DeliveryOptions(hasOrigin, worktreeMode)
		fmt.Println(ui.RenderDeliveryChoice(feature, opts))
		fmt.Println(gitutil.DeliveryDirective(feature, opts))
	} else {
		fmt.Println(ui.RenderStep("Next step", wf.CurrentStep))
	}
	if warn := workflow.ProductionReadinessWarning(feature, cfg); warn != "" {
		fmt.Println(ui.RenderProductionReadinessWarning(feature))
	}
	return nil
}
