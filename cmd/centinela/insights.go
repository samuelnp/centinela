package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	insightsTop  int
	insightsJSON bool
)

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Report governance-telemetry analytics (read-only)",
	Long: "Reads the append-only governance telemetry log and reports the " +
		"most-triggered blocks, most-failed gates, features with the most " +
		"rework, and the mean steps-to-green. Read-only: it never mutates or " +
		"writes telemetry. An empty or missing log prints a clean " +
		"\"no telemetry yet\" report and exits 0. Use --json for a stable, " +
		"machine-readable Report; --top N bounds each ranked section.",
	RunE:          runInsights,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	insightsCmd.Flags().IntVar(&insightsTop, "top", 5, "Max entries per ranked section")
	insightsCmd.Flags().BoolVar(&insightsJSON, "json", false, "Emit the structured Report as indented JSON")
	rootCmd.AddCommand(insightsCmd)
}

func runInsights(_ *cobra.Command, _ []string) error {
	events, err := telemetry.ReadDefault()
	if err != nil {
		return err
	}
	report := insights.Compute(events, insightsTop)
	if insightsJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	fmt.Fprintln(os.Stdout, ui.RenderInsights(report))
	return nil
}
