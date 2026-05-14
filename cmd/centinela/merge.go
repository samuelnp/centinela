package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <feature>",
	Short: "Merge a completed feature worktree back into the main branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(_ *cobra.Command, args []string) error {
	feature := args[0]
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}
	if conflicts := worktree.DetectSpecConflicts(".", feature); len(conflicts) > 0 {
		return fmt.Errorf("spec conflicts block merge: %s", worktree.FormatSpecConflicts(conflicts))
	}
	outcome, err := worktree.Merge(".", feature, runValidateForMerge)
	if err != nil {
		return err
	}
	if outcome.TextConflict || outcome.ValidateFail {
		fmt.Println(ui.RenderStep("Merge Steward required", feature))
		return fmt.Errorf("merge requires steward review — see %s", outcome.StewardHint())
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Merged %q into main and removed its worktree.", feature)))
	return nil
}

func runValidateForMerge(_ string) (bool, string) {
	if err := executeValidation(); err != nil {
		return false, err.Error()
	}
	return true, ""
}
