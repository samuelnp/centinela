package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/docgen"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

var (
	mergeContinue    bool
	mergeForceRemove bool
)

const mergePortalTitle = "Centinela Project Documentation"

// docsPortalRegen regenerates the documentation portal after a clean merge.
// It is a package-level seam so tests can swap it; production wiring calls
// docgen.Generate in the primary tree (docgen reads CWD-relative inputs and
// used to write into the worktree that was about to be deleted), and is
// best-effort because its inputs may be absent.
var docsPortalRegen = func(repo string) error {
	return inDir(repo, func() error {
		return docgen.Generate("docs/project-docs/index.html", mergePortalTitle)
	})
}

var mergeCmd = &cobra.Command{
	Use:   "merge <feature>",
	Short: "Merge a completed feature worktree back into the main branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func init() {
	mergeCmd.Flags().BoolVar(&mergeContinue, "continue", false, "Resume a stalled merge after the Merge Steward writes evidence")
	mergeCmd.Flags().BoolVar(&mergeForceRemove, "force-remove", false, "Retry worktree removal with --force after a verified merge (discards untracked files in the worktree)")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(_ *cobra.Command, args []string) error {
	feature := args[0]
	if err := worktree.ValidateFeatureSlug(feature); err != nil {
		return err
	}
	// Resolve the primary working tree ONCE and do everything there. From
	// inside a feature worktree, repo="." would make `git merge` self-no-op
	// and fabricate success, and would look for the pending marker in the
	// wrong tree (2048-rust retrospective, WS1.1). Never guess.
	repo, err := worktree.PrimaryTree(".")
	if err != nil {
		return err
	}
	if mergeContinue {
		return runMergeContinue(repo, feature)
	}
	if worktree.IsInsideWorktree(".") {
		fmt.Println(ui.RenderStep("Merging in primary working tree", repo))
	}
	if conflicts := worktree.DetectSpecConflicts(repo, feature); len(conflicts) > 0 {
		return fmt.Errorf("spec conflicts block merge: %s", worktree.FormatSpecConflicts(conflicts))
	}
	outcome, err := worktree.Merge(repo, feature, runValidateForMerge, mergeRemovalOpts()...)
	if err != nil {
		return reportMergeFailure(outcome, "centinela merge "+feature, err)
	}
	if outcome.TextConflict || outcome.ValidateFail {
		return dispatchSteward(repo, outcome)
	}
	// A clean merge supersedes any stale pending marker from a prior
	// stalled attempt, so the hook stops re-emitting its directive.
	if err := worktree.ClearPending(repo, feature); err != nil {
		return err
	}
	// Refresh the portal once per delivery. Best-effort: docgen needs
	// PROJECT.md/ROADMAP.md/roadmap.json, which may be absent, so a regen
	// failure must never fail an otherwise-clean merge.
	if err := docsPortalRegen(repo); err != nil {
		fmt.Printf("notice: portal regen skipped: %v\n", err)
	}
	return reportMergeSuccess(outcome)
}

func mergeRemovalOpts() []worktree.MergeOption {
	if mergeForceRemove {
		return []worktree.MergeOption{worktree.WithForceRemove()}
	}
	return nil
}
