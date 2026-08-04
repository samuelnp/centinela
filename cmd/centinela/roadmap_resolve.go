package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/roadmapstate"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve a conflicted roadmap.json by unioning both sides' Backlog findings",
	Args:  cobra.NoArgs,
	RunE:  runRoadmapResolve,
}

func init() {
	roadmapCmd.AddCommand(roadmapResolveCmd)
}

// gitStageRunner is the production seam: git in the current tree.
func gitStageRunner(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// runRoadmapResolve mechanizes the procedure every delivery ran by hand this
// week. It refuses BEFORE writing anything, so a real divergence leaves the
// conflict markers and the index exactly as git left them.
func runRoadmapResolve(_ *cobra.Command, _ []string) error {
	return resolveRoadmapState(".", gitStageRunner)
}

func resolveRoadmapState(dir string, run roadmap.StageRunner) error {
	jsonConflicted := roadmap.Conflicted(dir, roadmapstate.RoadmapJSON, run)
	mdConflicted := roadmap.Conflicted(dir, roadmapstate.MarkdownFile, run)
	if !jsonConflicted && !mdConflicted {
		fmt.Println(ui.RenderResolveNothingToDo())
		return nil
	}
	summary := ui.RenderResolveMarkdownOnly()
	if jsonConflicted {
		merged, err := mergeConflictedRoadmap(dir, run)
		if err != nil {
			return err
		}
		if err := roadmap.WriteRoadmapJSON(merged.Doc); err != nil {
			return err
		}
		summary = ui.RenderResolveSummary(merged)
	}
	if _, err := roadmap.RegenerateMarkdown(); err != nil {
		return err
	}
	if err := stageRoadmapState(dir, run); err != nil {
		return err
	}
	fmt.Println(summary)
	return nil
}

// mergeConflictedRoadmap reads the three index stages and merges them. Every
// failure returns before any file is written or staged.
func mergeConflictedRoadmap(dir string, run roadmap.StageRunner) (roadmap.Merged, error) {
	stages, ok, err := roadmap.ReadStages(dir, roadmapstate.RoadmapJSON, run)
	if err != nil {
		return roadmap.Merged{}, err
	}
	if !ok {
		return roadmap.Merged{}, fmt.Errorf("%s is not conflicted", roadmapstate.RoadmapJSON)
	}
	return roadmap.Resolve(stages.Base, stages.Ours, stages.Theirs)
}

// stageRoadmapState marks both roadmap-state paths resolved.
func stageRoadmapState(dir string, run roadmap.StageRunner) error {
	if out, err := run(dir, append([]string{"add", "--"}, roadmapstate.Paths()...)...); err != nil {
		return fmt.Errorf("staging resolved roadmap state: %v: %s", err, strings.TrimSpace(out))
	}
	return nil
}
