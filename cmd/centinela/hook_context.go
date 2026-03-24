package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var hookContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Hook: show active workflow status on every prompt",
	RunE:  runHookContext,
}

func init() {
	hookCmd.AddCommand(hookContextCmd)
}

func runHookContext(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE on large prompts
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	if r, err := roadmap.Load(); err == nil {
		fmt.Println(ui.RenderRoadmapSummary(r))
	}
	if len(entries) == 0 {
		fmt.Println(ui.StyleMuted.Render("No active workflows."))
		return nil
	}
	var wfs []*workflow.Workflow
	for _, path := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(path), ".json"))
		if err != nil {
			continue
		}
		wfs = append(wfs, wf)
	}
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	fmt.Println(ui.RenderContext(wfs))
	for _, wf := range wfs {
		if wf.CurrentStep != "done" && workflow.ValidateArtifacts(wf.Feature, wf.CurrentStep, cfg) == nil {
			fmt.Println(ui.RenderReviewReady(wf.Feature, wf.CurrentStep, nextStep(wf.CurrentStep)))
		}
	}
	for _, wf := range wfs {
		if wf.CurrentStep == "plan" && workflow.ValidateArtifacts(wf.Feature, "plan", cfg) != nil {
			if _, err := os.Stat(fmt.Sprintf("docs/features/%s.md", wf.Feature)); os.IsNotExist(err) {
				fmt.Println(ui.RenderFeatureBriefNeeded(wf.Feature))
			}
		}
	}
	for _, wf := range wfs {
		if wf.CurrentStep != "tests" {
			continue
		}
		path := fmt.Sprintf(".workflow/%s-edge-cases.md", wf.Feature)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(ui.RenderEdgeCaseReportNeeded(wf.Feature))
		}
	}
	return nil
}

func nextStep(current string) string {
	for i, s := range workflow.StepOrder {
		if s == current && i+1 < len(workflow.StepOrder) {
			return workflow.StepOrder[i+1]
		}
	}
	return "done"
}
