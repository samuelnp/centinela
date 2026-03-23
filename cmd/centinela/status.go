package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/workflow"
)

var statusCmd = &cobra.Command{
	Use:   "status <feature>",
	Short: "Show workflow status for a feature",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

var statusAllCmd = &cobra.Command{
	Use:   "status-all",
	Short: "Show all active workflows",
	RunE:  runStatusAll,
}

var statusRunner = runStatusModel

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(statusAllCmd)
}

func runStatus(_ *cobra.Command, args []string) error {
	wf, err := workflow.Load(args[0])
	if err != nil {
		return err
	}
	return statusRunner([]*workflow.Workflow{wf})
}

func runStatusAll(_ *cobra.Command, _ []string) error {
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	if len(entries) == 0 {
		fmt.Println("No active workflows.")
		return nil
	}
	var wfs []*workflow.Workflow
	for _, path := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(path), ".json"))
		if err != nil {
			continue
		}
		wfs = append(wfs, wf)
	}
	return statusRunner(wfs)
}
