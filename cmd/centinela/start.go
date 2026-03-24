package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var startCmd = &cobra.Command{
	Use:   "start <feature>",
	Short: "Start a new feature workflow",
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(_ *cobra.Command, args []string) error {
	feature := args[0]

	if _, err := os.Stat("PROJECT.md"); err != nil {
		return fmt.Errorf("PROJECT.md not found — run the /centinela-setup skill to initialise your project")
	}

	if _, err := os.Stat(workflow.FilePath(feature)); err == nil {
		return fmt.Errorf("workflow for %q already exists — use 'status' to check progress", feature)
	}
	order, err := workflowOrderForFeature(feature)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(workflow.WorkflowDir, 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", workflow.WorkflowDir, err)
	}

	wf := workflow.NewWithOrder(feature, order)
	if err := workflow.Save(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Workflow started for %q.", feature)))
	fmt.Println(ui.StyleMuted.Render("Steps: " + stepArrow(order)))
	fmt.Println(ui.RenderStep("Current step", "plan"))
	return nil
}

func stepArrow(order []string) string {
	if len(order) == 3 {
		return "plan → code → validate"
	}
	return "plan → code → tests → validate"
}
