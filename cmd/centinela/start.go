package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
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
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}

	if _, err := os.Stat("PROJECT.md"); err != nil {
		return fmt.Errorf("PROJECT.md not found — run the /centinela-setup skill to initialise your project")
	}

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}

	wtPath, err := worktree.MaybeProvision(".", feature, cfg)
	if err != nil {
		return err
	}
	if wtPath != "" {
		fmt.Println(ui.RenderSuccess("Worktree ready at " + wtPath))
	}

	if _, err := os.Stat(workflow.FilePath(feature)); err == nil {
		if wtPath != "" {
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Resuming existing workflow for %q.", feature)))
			return nil
		}
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
	wf.WorktreePath = wtPath
	if err := workflow.Save(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Workflow started for %q.", feature)))
	fmt.Println(ui.StyleMuted.Render("Steps: " + stepArrow(order)))
	fmt.Println(ui.RenderStep("Current step", "plan"))
	return nil
}

func stepArrow(order []string) string {
	return strings.Join(order, " → ")
}
