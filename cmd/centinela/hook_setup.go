package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
)

var hookSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Hook: prompt setup flow when project artifacts are missing",
	RunE:  runHookSetup,
}

func init() {
	hookCmd.AddCommand(hookSetupCmd)
}

func runHookSetup(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE

	if _, err := os.Stat("PROJECT.md.template"); err != nil {
		return nil // not a centinela project
	}
	if _, err := os.Stat("PROJECT.md"); err != nil {
		fmt.Println(ui.RenderSetupNeeded())
		return nil
	}
	if _, err := os.Stat("ROADMAP.md"); err != nil {
		fmt.Println(ui.RenderRoadmapNeeded())
		return nil
	}
	if _, err := os.Stat("docs/architecture/production-readiness-prompt.md"); err != nil {
		fmt.Println(ui.RenderProductionReadinessSetupNeeded())
		return nil
	}
	return nil
}
