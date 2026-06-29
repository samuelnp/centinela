package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/cost"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

var costJSON bool

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Report token spend vs budget per feature/step/model (read-only)",
	Long: "Folds the governance telemetry log into per-feature, per-step, and " +
		"per-model token spend and compares it to the [cost] budgets. Read-only " +
		"and soft: it never blocks. An empty log prints a clean \"no cost " +
		"samples yet\" line and exits 0. Use --json for a stable Report.",
	RunE:          runCost,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	costCmd.Flags().BoolVar(&costJSON, "json", false, "Emit the structured Report as indented JSON")
	rootCmd.AddCommand(costCmd)
}

func runCost(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	events, err := telemetry.ReadDefault()
	if err != nil {
		return err
	}
	report := cost.Build(cost.Fold(events), cfg.Cost)
	if costJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(ui.RenderCost(report))
	return nil
}
