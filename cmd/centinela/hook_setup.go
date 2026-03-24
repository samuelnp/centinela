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

	hasTemplate := exists("PROJECT.md.template")
	hasProject := exists("PROJECT.md")
	if !hasTemplate && !hasProject && !exists("centinela.toml") {
		return nil
	}
	if !hasProject {
		fmt.Println("CENTINELA DIRECTIVE: setup required. Ask setup questions and write PROJECT.md.")
		fmt.Println(ui.RenderSetupNeeded())
		return nil
	}
	if !exists("ROADMAP.md") {
		fmt.Println("CENTINELA DIRECTIVE: roadmap required. Define roadmap before feature work.")
		fmt.Println(ui.RenderRoadmapNeeded())
		return nil
	}
	if !exists("docs/architecture/production-readiness-prompt.md") {
		fmt.Println("CENTINELA DIRECTIVE: configure production-readiness prompt before continuing.")
		fmt.Println(ui.RenderProductionReadinessSetupNeeded())
		return nil
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
