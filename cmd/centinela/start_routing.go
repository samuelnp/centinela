package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// printRoutingDirective emits the dynamic-mode routing line for the workflow's
// first step, so the orchestrator decides tiers BEFORE delegating — start is the
// sanctioned routing window (no evidence exists yet, so no downgrade is refused).
//
// Under the default static mode it returns before touching anything, leaving
// `centinela start` output byte-identical to today's.
func printRoutingDirective(wf *workflow.Workflow, cfg *config.Config) {
	if !config.DynamicRoutingEnabled(cfg) {
		return
	}
	models, _ := orchestrationRouting(cfg)
	line := orchestration.RoutingDirective(wf.Feature,
		workflow.RequiredEvidenceRoles(wf.Feature, wf.CurrentStep),
		workflow.RouteTiers(wf), config.OrchestrationFloors(cfg), models)
	if line == "" {
		return
	}
	fmt.Printf("CENTINELA DIRECTIVE: %s\n", line)
}
