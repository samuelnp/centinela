package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	backlogStaleOnly bool
	backlogOlderThan int
	backlogJSON      bool
)

var roadmapBacklogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "List deferred Backlog findings oldest first, with their age",
	Args:  cobra.NoArgs,
	RunE:  runRoadmapBacklog,
}

func init() {
	roadmapBacklogCmd.Flags().BoolVar(&backlogStaleOnly, "stale", false,
		"show only findings older than the threshold")
	roadmapBacklogCmd.Flags().IntVar(&backlogOlderThan, "older-than", 0,
		"override the staleness threshold in days (default 14)")
	roadmapBacklogCmd.Flags().BoolVar(&backlogJSON, "json", false,
		"emit the machine shape for every finding")
	roadmapCmd.AddCommand(roadmapBacklogCmd)
}

// runRoadmapBacklog is a thin orchestrator: the aging, ordering and counting
// all live in internal/roadmap, the table in internal/ui.
func runRoadmapBacklog(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	threshold := backlogThreshold()
	aged := roadmap.AgeBacklog(r, time.Now(), threshold)
	stats := roadmap.SummarizeBacklog(aged, threshold)
	rows := aged
	if backlogStaleOnly {
		rows = roadmap.StaleOnly(aged)
	}
	if backlogJSON {
		return emitBacklogJSON(rows, stats)
	}
	fmt.Println(ui.RenderBacklogList(rows, stats, backlogStaleOnly))
	return nil
}

// backlogThreshold resolves --older-than over the shipped default.
func backlogThreshold() int {
	if backlogOlderThan > 0 {
		return backlogOlderThan
	}
	return roadmap.DefaultStaleDays
}

// printBacklogNudge prints the completion prompt when the roadmap is finished
// and the Backlog is not. Any load failure suppresses it silently: a hint must
// never break `complete`.
func printBacklogNudge() {
	r, err := roadmap.Load()
	if err != nil {
		return
	}
	if n, ok := roadmap.NudgeFor(r, time.Now(), roadmap.DefaultStaleDays); ok {
		fmt.Println(ui.RenderBacklogNudge(n))
	}
}
