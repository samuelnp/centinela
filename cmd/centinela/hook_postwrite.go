package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var hookPostwriteCmd = &cobra.Command{
	Use:   "postwrite",
	Short: "PostToolUse hook: inject workflow tag after every file write",
	RunE:  runHookPostwrite,
}

func init() {
	hookCmd.AddCommand(hookPostwriteCmd)
}

func runHookPostwrite(_ *cobra.Command, _ []string) error {
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	for _, path := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(path), ".json"))
		if err != nil {
			continue
		}
		fmt.Println(ui.RenderTag(wf))
	}
	return nil
}
