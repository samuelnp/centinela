package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var hookSessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Hook: rehydrate roadmap context on session entry (SessionStart)",
	RunE:  runHookSession,
}

func init() {
	hookCmd.AddCommand(hookSessionCmd)
}

func runHookSession(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE
	emitUpdateNotice()
	r, err := roadmap.Load()
	if err != nil || r == nil {
		// No roadmap (absent or invalid) — exit silently, no payload.
		return nil
	}
	// Compute the ready set + incomplete flag here so the renderer stays pure.
	ready := roadmap.ReadySet(r)
	planned, inProgress, _ := r.Summary()
	hasIncomplete := planned > 0 || inProgress > 0
	fmt.Println("CENTINELA DIRECTIVE: session rehydration — recovered project state below.")
	fmt.Println(ui.RenderSessionRehydration(r, ready, hasIncomplete))
	return nil
}

// emitUpdateNotice prints the throttled, fail-silent "update available" notice.
// It never blocks or errors the session: a "" notice (current, dev build,
// within-TTL cache, or any network/parse error) prints nothing.
func emitUpdateNotice() {
	if notice := newSelfUpdater(Version).Notice(); notice != "" {
		fmt.Println(notice)
	}
}
