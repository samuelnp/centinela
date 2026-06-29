package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/cost"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

// emitCostWarning prints the cost-governance soft-gate line when the active
// feature/step is over its [cost] budget. It is a pure side-effect: no-op when
// cost is inactive, there is no active feature, or nothing is over budget, and
// it NEVER affects validate's exit code (cost is a soft gate).
func emitCostWarning(cfg *config.Config) {
	if cfg == nil || !cfg.Cost.IsActive() {
		return
	}
	wf := activeWorkflow(mustGetwd())
	if wf == nil {
		return
	}
	events, err := telemetry.ReadDefault()
	if err != nil {
		return
	}
	st, over := cost.ActiveStatus(cost.Fold(events), cfg.Cost, wf.Feature, wf.CurrentStep)
	if !over {
		return
	}
	fmt.Println()
	fmt.Println(ui.RenderCostWarning(st))
}
