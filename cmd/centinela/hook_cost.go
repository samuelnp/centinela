package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/cost"
	"github.com/samuelnp/centinela/internal/telemetry"
)

var hookCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Hook: attribute host-harness transcript token spend to the active feature/step",
	RunE:  runHookCost,
}

func init() {
	hookCmd.AddCommand(hookCostCmd)
}

type costHookInput struct {
	Cwd            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
}

// runHookCost reads the harness transcript delta since the saved cursor and
// records it as a cost-sample for the active feature/step. Every failure mode
// (cost disabled, no transcript, no active feature, parse error) is a silent
// no-op — the hook must never block or error the host command.
func runHookCost(_ *cobra.Command, _ []string) error {
	raw, _ := io.ReadAll(os.Stdin)
	var in costHookInput
	_ = json.Unmarshal(raw, &in)
	if in.TranscriptPath == "" {
		return nil
	}
	cfg, err := config.Load()
	if err != nil || !cfg.Cost.IsActive() {
		return nil
	}
	cwd := in.Cwd
	if cwd == "" {
		cwd = mustGetwd()
	}
	wf := activeWorkflow(cwd)
	if wf == nil {
		return nil
	}
	offset := cost.LoadCursor().OffsetFor(in.TranscriptPath)
	tokIn, tokOut, newOffset, _ := cost.SumFrom(in.TranscriptPath, offset)
	telemetry.RecordCostSample(cfg, wf.Feature, wf.CurrentStep, resolveEmitModel(wf, cfg), tokIn, tokOut)
	cost.SaveCursor(in.TranscriptPath, newOffset)
	return nil
}
