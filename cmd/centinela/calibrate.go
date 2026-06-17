package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/calibration"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

var calibrateJSON bool

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "Recommend per-model enforcement profiles from telemetry (read-only)",
	Long: "Reads the append-only governance telemetry log, groups events per " +
		"driver model, measures friction (rework per successful advance), and " +
		"recommends a tighter, looser, or unchanged enforcement profile for each " +
		"model — every recommendation backed by the raw counts that drove it. " +
		"Read-only and advisory: it never mutates telemetry or writes config. An " +
		"empty or missing log prints a clean \"no telemetry yet\" report and " +
		"exits 0. Use --json for a stable, machine-readable Report.",
	RunE:          runCalibrate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	calibrateCmd.Flags().BoolVar(&calibrateJSON, "json", false, "Emit the structured Report as indented JSON")
	rootCmd.AddCommand(calibrateCmd)
}

func runCalibrate(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	events, err := telemetry.ReadDefault()
	if err != nil {
		return err
	}
	report := calibration.Calibrate(events, cfg)
	if calibrateJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	fmt.Fprintln(os.Stdout, ui.RenderCalibration(report))
	return nil
}
