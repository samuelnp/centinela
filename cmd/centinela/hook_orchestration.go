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
	if cfg, err := config.Load(); err == nil {
		models, modelMap = orchestrationRouting(cfg)
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
		names, files, tiers := annotateRoles(wf.Feature, roles, models, modelMap)
		fmt.Printf("CENTINELA DIRECTIVE: orchestrator only for %q/%q; delegate to [%s].\n", wf.Feature, wf.CurrentStep, strings.Join(names, ", "))
		fmt.Printf("Required evidence before centinela complete %s: %s\n", wf.Feature, strings.Join(files, ", "))
		fmt.Printf("CENTINELA DIRECTIVE: model reference: %s\n", orchestration.ModelReference(tiers))
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
