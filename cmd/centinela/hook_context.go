package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
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
	cfg, err := config.Load()
	if err != nil {
		// Hooks must never break the host session: degrade to defaults and
		// surface the failure in the injected context instead.
		fmt.Println("config warning: " + err.Error())
		cfg = &config.Config{}
	}
	wfs := loadActiveWorkflows()
	if r, err := roadmap.Load(); err == nil {
		fmt.Println(ui.RenderRoadmapSummary(r))
	}
	if len(wfs) == 0 {
		fmt.Println("CENTINELA DIRECTIVE: no active workflow. Start a feature before implementation.")
		fmt.Println(ui.RenderSuccess("No active workflows."))
		return nil
	}
	const activePanelCap = 5
	shown, more := workflow.CapActive(wfs, activePanelCap)
	fmt.Println(ui.RenderContextCapped(shown, more))
	for _, wf := range wfs {
		if shouldRenderReviewPrompt(wf, cfg) && workflow.ValidateArtifacts(wf.Feature, wf.CurrentStep, cfg) == nil {
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
		if orchestration.IsUserFacingFeature(wf.Feature) {
			if _, err := os.Stat("docs/project-docs/index.html"); os.IsNotExist(err) {
				fmt.Println(ui.RenderDocumentationNeeded(wf.Feature))
			}
			continue
		}
		if _, err := os.Stat(".workflow/" + wf.Feature + "-changelog.md"); os.IsNotExist(err) {
			fmt.Println(ui.RenderChangelogNeeded(wf.Feature))
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
