package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/teamdashboard"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var dashboardJSON bool

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Print a read-only multi-feature team-status board",
	Long: "Aggregates current Centinela state into three panels — in-flight " +
		"features, roadmap burn-down, and gate health — from the active " +
		"workflow JSONs, roadmap.json, and the telemetry event log. Read-only: " +
		"it never writes or mutates any file. Missing or empty sources each " +
		"render an honest empty-state panel, never an error. Owners are a " +
		"best-effort git-derived column. Use --json for the stable Dashboard.",
	RunE:          runDashboard,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardJSON, "json", false, "Emit the structured Dashboard as indented JSON")
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(_ *cobra.Command, _ []string) error {
	active := workflow.ActiveWorkflows(workflow.WorkflowDir)
	rm, _ := roadmap.Load() // ignore error -> nil -> honest empty roadmap panel
	events, err := telemetry.ReadDefault()
	if err != nil {
		return err
	}
	dash := teamdashboard.Compute(teamdashboard.Inputs{
		Active:  active,
		Roadmap: rm,
		Events:  events,
		Owners:  dashboardOwners(active),
		Now:     time.Now().UTC(),
	})
	if dashboardJSON {
		data, err := json.MarshalIndent(dash, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	fmt.Fprintln(os.Stdout, ui.RenderDashboard(dash))
	return nil
}
