package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
)

// routeCmd is the thin orchestrator for `centinela route`. Every decision —
// floors, tier ordering, the refusal matrix, step scheduling — lives in
// internal/orchestration and internal/workflow per G7. This file ONLY wires
// subcommands and owns the shared dynamic-mode guard (refusal matrix rule 1).
var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "Route a subagent role to a model tier for one feature (dynamic mode)",
	Long: "Per-feature model routing. Requires [orchestration] routing_mode =\n" +
		"\"dynamic\"; under the default static mode every subcommand refuses and\n" +
		"model resolution stays exactly as configured project-wide.",
}

func init() {
	rootCmd.AddCommand(routeCmd)
}

// loadRoutingConfig loads the config and applies refusal matrix rule 1: routing
// commands exist only under dynamic mode, and the error names the key to set.
func loadRoutingConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if !config.DynamicRoutingEnabled(cfg) {
		return nil, fmt.Errorf("dynamic routing is off — set [orchestration] routing_mode = %q in centinela.toml", config.RoutingModeDynamic)
	}
	return cfg, nil
}
