package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var (
	reviseTo     string
	reviseReason string
)

var reviseCmd = &cobra.Command{
	Use:   "revise <feature>",
	Short: "Rewind a feature to an earlier step, shedding downstream evidence",
	Long: "Perform a controlled backward transition: re-open every step after " +
		"--to to pending, set --to in-progress, and delete only the re-opened " +
		"steps' certification evidence (.workflow/<feature>-*) so the next " +
		"complete re-runs their gates. Source, test, and docs files are untouched.",
	Args: cobra.ExactArgs(1),
	RunE: runRevise,
}

func init() {
	reviseCmd.Flags().StringVar(&reviseTo, "to", "", "earlier step to rewind to (required)")
	reviseCmd.Flags().StringVar(&reviseReason, "reason", "", "why the rewind is needed (required)")
	_ = reviseCmd.MarkFlagRequired("to")
	_ = reviseCmd.MarkFlagRequired("reason")
	rootCmd.AddCommand(reviseCmd)
}

func runRevise(_ *cobra.Command, args []string) error {
	feature := args[0]
	if strings.TrimSpace(reviseReason) == "" {
		return fmt.Errorf("--reason must not be empty")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}

	from := wf.CurrentStep
	reopened, err := wf.RewindTo(reviseTo, reviseReason)
	if err != nil {
		return err
	}
	count, err := invalidateDownstream(feature, reopened)
	if err != nil {
		return err
	}
	if err := saveWorkflow(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	telemetry.RecordRevised(cfg, feature, from, reviseTo, resolveEmitModel(wf, cfg))

	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Revised %q from %q to %q (%d evidence artifacts invalidated).",
		feature, from, reviseTo, count)))
	fmt.Println(ui.RenderStep("Current step", wf.CurrentStep))
	return nil
}
