package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

var deferSummary string
var deferSource string

var roadmapDeferCmd = &cobra.Command{
	Use:   "defer <slug>",
	Short: "Capture an out-of-scope finding into the validate-exempt Backlog phase",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapDefer,
}

func init() {
	roadmapDeferCmd.Flags().StringVar(&deferSummary, "summary", "", "one-line finding summary (required)")
	roadmapDeferCmd.Flags().StringVar(&deferSource, "source", "", "provenance as <feature>/<role> (auto-resolved from worktree CWD)")
	roadmapCmd.AddCommand(roadmapDeferCmd)
}

func runRoadmapDefer(_ *cobra.Command, args []string) error {
	src := resolveDeferSource(deferSource)
	if err := roadmap.Defer(roadmap.RoadmapFile, roadmap.DeferOptions{
		Slug:    args[0],
		Summary: deferSummary,
		Source:  src,
	}); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Deferred %q to the Backlog phase.", args[0])))
	return nil
}

// resolveDeferSource parses an explicit --source override (<feature>/<role>) or
// auto-detects the feature from the worktree CWD. Returns nil at repo root.
func resolveDeferSource(flag string) *roadmap.Source {
	if flag != "" {
		feature, role, _ := strings.Cut(flag, "/")
		return &roadmap.Source{Feature: feature, Role: role}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	if feature, _ := worktree.DetectFeatureFromCwd(cwd); feature != "" {
		return &roadmap.Source{Feature: feature}
	}
	return nil
}
