package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
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
	// Hooks must never break the host session: on a config error warn and
	// fall back to defaults rather than abort.
	models := orchestration.RoleModels{}
	modelMap := orchestration.ModelMap{}
	// dynamic/floors stay zero under the default static mode, so every branch
	// they guard below is skipped and the emitted bytes are exactly today's.
	dynamic := false
	var floors map[string]string
	if cfg, err := config.Load(); err == nil {
		models, modelMap = orchestrationRouting(cfg)
		if config.DynamicRoutingEnabled(cfg) {
			dynamic, floors = true, config.OrchestrationFloors(cfg)
		}
	} else {
		fmt.Println("config warning: " + err.Error())
	}
	for _, wf := range loadActiveWorkflows() {
		if wf.OrchestrationMode != workflow.StrictOrchestrationMode {
			continue
		}
		// Contract-aware: the directive must name the SAME roles `complete`
		// enforces, or a legacy workflow is told to write evidence the gate
		// will then refuse.
		roles := workflow.RequiredEvidenceRoles(wf.Feature, wf.CurrentStep)
		if len(roles) == 0 {
			continue
		}
		// Per-workflow overlay: routes belong to ONE feature, so the shared
		// models map must never be mutated across the loop. Floors travel INTO
		// the overlay: state is agent-writable, so the floor must bind where the
		// model is resolved, not only where a route is recorded.
		wfModels := models
		if dynamic {
			wfModels = orchestration.ApplyRoutes(models, workflow.RouteTiers(wf), floors)
		}
		names, files, tiers := annotateRoles(wf.Feature, roles, wfModels, modelMap)
		fmt.Printf("CENTINELA DIRECTIVE: orchestrator only for %q/%q; delegate to [%s].\n", wf.Feature, wf.CurrentStep, strings.Join(names, ", "))
		fmt.Printf("Required evidence before centinela complete %s: %s\n", wf.Feature, strings.Join(files, ", "))
		fmt.Printf("CENTINELA DIRECTIVE: model reference: %s\n", orchestration.ModelReference(tiers))
		if dynamic {
			// models (not wfModels) is the static reference the decision is
			// measured against; the line disappears once every role is routed.
			if line := orchestration.RoutingDirective(wf.Feature, roles, workflow.RouteTiers(wf), floors, models); line != "" {
				fmt.Printf("CENTINELA DIRECTIVE: %s\n", line)
			}
		}
		// Only a workflow pinned to adversarial-v1 is told to spawn the
		// verifier; a legacy one would be pointed at a role its gate refuses.
		if contract := orchestration.DelegationContract(wf.CurrentStep); contract != "" && wf.UsesAdversarialVerifier() {
			fmt.Printf("CENTINELA DIRECTIVE: %s\n", contract)
		}
	}
	return nil
}

// orchestrationRouting maps the config-leaf routing tables onto the domain
// resolver types. Pure data shuffling — no decision logic (G7).
func orchestrationRouting(cfg *config.Config) (orchestration.RoleModels, orchestration.ModelMap) {
	models := orchestration.RoleModels{}
	for role, tier := range config.OrchestrationModelTiers(cfg) {
		models[role] = orchestration.RoleModel{Tier: tier}
	}
	for role, overrides := range config.OrchestrationModelOverrides(cfg) {
		models[role] = orchestration.RoleModel{Overrides: overrides}
	}
	return models, orchestration.ModelMap(config.OrchestrationModelMap(cfg))
}
