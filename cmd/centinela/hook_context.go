package main

import (
	"fmt"
	"io"
	"os"

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
	wfs := loadActiveWorkflows()
	if r, err := roadmap.Load(); err == nil {
		fmt.Println(ui.RenderRoadmapSummary(r))
	}
	if len(wfs) == 0 {
		fmt.Println("CENTINELA DIRECTIVE: no active workflow. Start a feature before implementation.")
		fmt.Println(ui.RenderSuccess("No active workflows."))
		return nil
	}
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	fmt.Println(ui.RenderContext(wfs))
	for _, wf := range wfs {
		if wf.CurrentStep != "done" && workflow.ValidateArtifacts(wf.Feature, wf.CurrentStep, cfg) == nil {
			fmt.Println(ui.RenderReviewReady(wf.Feature, wf.CurrentStep, nextStepFor(wf, wf.CurrentStep)))
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
	for _, wf := range wfs {
		if wf.CurrentStep != "docs" {
			continue
		}
		if _, err := os.Stat("docs/project-docs/index.html"); os.IsNotExist(err) {
			fmt.Println(ui.RenderDocumentationNeeded(wf.Feature))
		}
	}
	return nil
}

func nextStep(current string) string {
	return nextStepFor(&workflow.Workflow{}, current)
}

func nextStepFor(wf *workflow.Workflow, current string) string {
	steps := wf.OrderedSteps()
	for i, s := range steps {
		if s == current && i+1 < len(steps) {
			return steps[i+1]
		}
	}
	return "done"
}
