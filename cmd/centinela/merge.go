package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/docgen"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

var mergeContinue bool

const mergePortalTitle = "Centinela Project Documentation"

// docsPortalRegen regenerates the documentation portal after a clean merge.
// It is a package-level seam so tests can swap it; production wiring calls
// docgen.Generate, which is best-effort (its inputs may be absent).
var docsPortalRegen = func() error {
	return docgen.Generate("docs/project-docs/index.html", mergePortalTitle)
}

var mergeCmd = &cobra.Command{
	Use:   "merge <feature>",
	Short: "Merge a completed feature worktree back into the main branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func init() {
	mergeCmd.Flags().BoolVar(&mergeContinue, "continue", false, "Resume a stalled merge after the Merge Steward writes evidence")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(_ *cobra.Command, args []string) error {
	feature := args[0]
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}
	if mergeContinue {
		return runMergeContinue(feature)
	}
	if conflicts := worktree.DetectSpecConflicts(".", feature); len(conflicts) > 0 {
		return fmt.Errorf("spec conflicts block merge: %s", worktree.FormatSpecConflicts(conflicts))
	}
	outcome, err := worktree.Merge(".", feature, runValidateForMerge)
	if err != nil {
		return err
	}
	if outcome.TextConflict || outcome.ValidateFail {
		return dispatchSteward(outcome)
	}
	// A clean merge supersedes any stale pending marker from a prior
	// stalled attempt, so the hook stops re-emitting its directive.
	if err := worktree.ClearPending(".", feature); err != nil {
		return err
	}
	// Refresh the portal once per delivery. Best-effort: docgen needs
	// PROJECT.md/ROADMAP.md/roadmap.json, which may be absent, so a regen
	// failure must never fail an otherwise-clean merge.
	if err := docsPortalRegen(); err != nil {
		fmt.Printf("notice: portal regen skipped: %v\n", err)
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
