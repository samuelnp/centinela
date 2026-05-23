package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/roadmapcheckpoint"
	"github.com/samuelnp/centinela/internal/ui"
)

var hookSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Hook: prompt setup flow when project artifacts are missing",
	RunE:  runHookSetup,
}

func init() {
	hookCmd.AddCommand(hookSetupCmd)
}

func runHookSetup(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE

	hasTemplate := exists("PROJECT.md.template")
	hasProject := exists("PROJECT.md")
	if !hasTemplate && !hasProject && !exists("centinela.toml") {
		return nil
	}
	if !hasProject {
		fmt.Println("CENTINELA DIRECTIVE: setup required. Ask setup questions and write PROJECT.md.")
		fmt.Println(ui.RenderSetupNeeded())
		return nil
	}
	if !exists("ROADMAP.md") {
		fmt.Println("CENTINELA DIRECTIVE: roadmap required. Define roadmap before feature work.")
		fmt.Println(ui.RenderRoadmapNeeded())
		return nil
	}
	r, err := roadmap.Load()
	if err != nil {
		fmt.Println(roadmapJSONDirective(err))
		fmt.Println(ui.RenderRoadmapJSONNeeded(err))
		return nil
	}
	if !exists(".workflow/roadmap-analysis.md") || !exists(".workflow/roadmap-analysis.json") {
		fmt.Println("CENTINELA DIRECTIVE: roadmap analysis required. Delegate to senior product manager.")
		fmt.Println(ui.RenderRoadmapAnalysisNeeded())
		return nil
	}
	if !exists(".workflow/roadmap-quality.md") || !exists(".workflow/roadmap-quality.json") {
		fmt.Println("CENTINELA DIRECTIVE: roadmap quality required. Delegate to roadmap quality evaluator.")
		fmt.Println(ui.RenderRoadmapQualityNeeded())
		return nil
	}
	if !exists("docs/architecture/production-readiness-prompt.md") {
		fmt.Println("CENTINELA DIRECTIVE: configure production-readiness prompt before continuing.")
		fmt.Println(ui.RenderProductionReadinessSetupNeeded())
		return nil
	}
	emitRoadmapCheckpoint(r)
	return nil
}

// emitRoadmapCheckpoint prints the checkpoint directive when the decision
// package says to emit. It never blocks: all rules live in roadmapcheckpoint.
func emitRoadmapCheckpoint(r *roadmap.Roadmap) {
	first, hasFirst := roadmapcheckpoint.FirstIncompleteBootstrap(r)
	d := roadmapcheckpoint.Decide(time.Now(), first, hasFirst, roadmapcheckpoint.NewOSFS())
	if d == roadmapcheckpoint.DecisionSuppressed {
		return
	}
	fmt.Println("CENTINELA DIRECTIVE: roadmap checkpoint. Ask the user to iterate or start the first Phase 0 feature.")
	fmt.Println(ui.RenderRoadmapCheckpoint(first))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func roadmapJSONDirective(err error) string {
	if os.IsNotExist(err) {
		return "CENTINELA DIRECTIVE: roadmap json required. Write .workflow/roadmap.json."
	}
	return "CENTINELA DIRECTIVE: roadmap json invalid. Fix .workflow/roadmap.json."
}
