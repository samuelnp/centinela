package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

var hookOrchestrationCmd = &cobra.Command{
	Use:   "orchestration",
	Short: "Hook: enforce strict step subagent delegation",
	RunE:  runHookOrchestration,
}

func init() {
	hookCmd.AddCommand(hookOrchestrationCmd)
}

func runHookOrchestration(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck
	for _, wf := range loadActiveWorkflows() {
		if wf.OrchestrationMode != workflow.StrictOrchestrationMode {
			continue
		}
		roles := orchestration.RequiredRolesForFeature(wf.Feature, wf.CurrentStep)
		if len(roles) == 0 {
			continue
		}
		files := []string{}
		names := []string{}
		for _, role := range roles {
			names = append(names, string(role))
			files = append(files, orchestration.MarkdownPath(wf.Feature, role))
			files = append(files, orchestration.JSONPath(wf.Feature, role))
		}
		fmt.Printf("CENTINELA DIRECTIVE: orchestrator only for %q/%q; delegate to [%s].\n", wf.Feature, wf.CurrentStep, strings.Join(names, ", "))
		fmt.Printf("Required evidence before centinela complete %s: %s\n", wf.Feature, strings.Join(files, ", "))
	}
	return nil
}
