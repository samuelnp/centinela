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
	r, err := roadmap.Load()
	if err != nil || r == nil {
		// No roadmap (absent or invalid) — exit silently, no payload.
		return nil
	}
	next, hasNext := roadmap.FirstIncomplete(r)
	fmt.Println("CENTINELA DIRECTIVE: session rehydration — recovered project state below.")
	fmt.Println(ui.RenderSessionRehydration(r, next, hasNext))
	return nil
}
