package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var promotePhase string
var promoteSummary string
var promoteScores string

var roadmapPromoteCmd = &cobra.Command{
	Use:   "promote <slug>",
	Short: "Promote a Backlog finding into a real phase via a quality-evaluator pass",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoadmapPromote,
}

func init() {
	roadmapPromoteCmd.Flags().StringVar(&promotePhase, "phase", "", "target non-Backlog phase (required)")
	roadmapPromoteCmd.Flags().StringVar(&promoteSummary, "summary", "", "override summary for the quality entry")
	roadmapPromoteCmd.Flags().StringVar(&promoteScores, "scores", "", "CSV ac,uv,dc,dep,ee,overall (each 1-10, overall >= 9)")
	roadmapCmd.AddCommand(roadmapPromoteCmd)
}

func runRoadmapPromote(cmd *cobra.Command, args []string) error {
	slug := args[0]
	if promotePhase == "" {
		return fmt.Errorf("--phase is required")
	}
	// Distinguish an unset --scores (evaluator/print-context path) from an
	// explicitly-empty --scores="" (a usage error): cobra's Changed reports
	// whether the flag was provided at all, regardless of its value.
	scoresSet := cmd.Flags().Changed("scores")
	if scoresSet && promoteScores == "" {
		return fmt.Errorf("--scores was set but empty; provide six comma-separated integers (ac,uv,dc,dep,ee,overall) or omit --scores for the evaluator context")
	}
	if !scoresSet {
		return printEvaluatorContext(slug)
	}
	return promoteScored(slug)
}

func printEvaluatorContext(slug string) error {
	finding, err := roadmap.LoadBacklogFinding(roadmap.RoadmapFile, slug)
	if err != nil {
		return err
	}
	fmt.Println(ui.RenderPromoteEvaluatorContext(finding, promotePhase))
	return nil
}

func promoteScored(slug string) error {
	scores, err := roadmap.ParseScores(promoteScores)
	if err != nil {
		return err
	}
	if _, err := roadmap.Promote(roadmap.RoadmapFile, roadmap.PromoteRequest{
		Slug: slug, Phase: promotePhase, Summary: promoteSummary, Scores: scores,
	}); err != nil {
		return err
	}
	return reportPromoteResult(slug)
}

func reportPromoteResult(slug string) error {
	r, err := roadmap.Load()
	if err != nil {
		return fmt.Errorf("promote wrote files but roadmap reload failed: %w", err)
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		return fmt.Errorf("promote wrote files but validate failed: %w", err)
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		return fmt.Errorf("promote wrote files but validate failed: %w", err)
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Promoted %q to %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", slug, promotePhase)))
	return nil
}
