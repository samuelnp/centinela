package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var hookContextCmd = &cobra.Command{
	Use:   "context",
	Short: "UserPromptSubmit hook: show active workflow status on every prompt",
	RunE:  runHookContext,
}

func init() {
	hookCmd.AddCommand(hookContextCmd)
}

func runHookContext(_ *cobra.Command, _ []string) error {
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
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
	fmt.Println(ui.RenderContext(wfs))
	return nil
}
